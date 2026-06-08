//go:build cgo

package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

type miniLMInputKind int

const (
	miniLMInputIDs miniLMInputKind = iota
	miniLMAttentionMask
	miniLMTokenTypeIDs
)

type miniLMInput struct {
	name string
	kind miniLMInputKind
}

type onnxMiniLMProvider struct {
	mu       sync.Mutex
	embedder *onnxMiniLMEmbedder
}

type onnxMiniLMEmbedder struct {
	mu        sync.Mutex
	session   *ort.DynamicAdvancedSession
	tokenizer *wordPieceTokenizer
	inputs    []miniLMInput
}

var ortInitMu sync.Mutex

func (p *onnxMiniLMProvider) Embed(ctx context.Context, text string, allowDownload bool) ([]float32, bool, error) {
	if strings.TrimSpace(text) == "" {
		return nil, false, nil
	}
	embedder, err := p.get(ctx, allowDownload)
	if err != nil {
		return nil, false, err
	}
	return embedder.Embed(text)
}

func (p *onnxMiniLMProvider) get(ctx context.Context, allowDownload bool) (*onnxMiniLMEmbedder, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.embedder != nil {
		return p.embedder, nil
	}
	assets, err := ensureSemanticAssets(ctx, allowDownload)
	if err != nil {
		return nil, err
	}
	if err := initializeONNXRuntime(assets.RuntimeLibPath); err != nil {
		return nil, fmt.Errorf("%w: initialize onnxruntime: %w", ErrSemanticUnavailable, err)
	}
	tokenizer, err := loadWordPieceTokenizer(assets.VocabPath)
	if err != nil {
		return nil, fmt.Errorf("%w: load tokenizer vocab: %w", ErrSemanticUnavailable, err)
	}
	inputInfos, outputInfos, err := ort.GetInputOutputInfo(assets.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("%w: inspect MiniLM ONNX model: %w", ErrSemanticUnavailable, err)
	}
	inputs, err := selectMiniLMInputs(inputInfos)
	if err != nil {
		return nil, fmt.Errorf("%w: select MiniLM inputs: %w", ErrSemanticUnavailable, err)
	}
	outputName, err := selectMiniLMOutput(outputInfos)
	if err != nil {
		return nil, fmt.Errorf("%w: select MiniLM output: %w", ErrSemanticUnavailable, err)
	}
	inputNames := make([]string, len(inputs))
	for i, input := range inputs {
		inputNames[i] = input.name
	}
	session, err := ort.NewDynamicAdvancedSession(assets.ModelPath, inputNames, []string{outputName}, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: load MiniLM ONNX model: %w", ErrSemanticUnavailable, err)
	}
	p.embedder = &onnxMiniLMEmbedder{
		session:   session,
		tokenizer: tokenizer,
		inputs:    inputs,
	}
	return p.embedder, nil
}

func initializeONNXRuntime(libPath string) error {
	ortInitMu.Lock()
	defer ortInitMu.Unlock()
	if ort.IsInitialized() {
		return nil
	}
	ort.SetSharedLibraryPath(libPath)
	return ort.InitializeEnvironment()
}

func selectMiniLMInputs(infos []ort.InputOutputInfo) ([]miniLMInput, error) {
	inputs := make([]miniLMInput, 0, len(infos))
	seen := make(map[miniLMInputKind]bool)
	for _, info := range infos {
		var kind miniLMInputKind
		switch info.Name {
		case "input_ids":
			kind = miniLMInputIDs
		case "attention_mask":
			kind = miniLMAttentionMask
		case "token_type_ids":
			kind = miniLMTokenTypeIDs
		default:
			return nil, fmt.Errorf("unsupported MiniLM input %q", info.Name)
		}
		inputs = append(inputs, miniLMInput{name: info.Name, kind: kind})
		seen[kind] = true
	}
	if !seen[miniLMInputIDs] || !seen[miniLMAttentionMask] {
		return nil, errors.New("MiniLM model missing input_ids or attention_mask")
	}
	return inputs, nil
}

func selectMiniLMOutput(infos []ort.InputOutputInfo) (string, error) {
	for _, info := range infos {
		if info.Name == "last_hidden_state" {
			return info.Name, nil
		}
	}
	for _, info := range infos {
		dims := info.Dimensions
		if len(dims) >= 3 && dims[len(dims)-1] == semanticEmbeddingDims {
			return info.Name, nil
		}
	}
	return "", errors.New("MiniLM model missing 384-dim token embedding output")
}

func (e *onnxMiniLMEmbedder) Embed(text string) ([]float32, bool, error) {
	ids, mask, tokenTypes, ok := e.tokenizer.Encode(text)
	if !ok {
		return nil, false, nil
	}
	shape := ort.NewShape(1, int64(len(ids)))
	inputValues := make([]ort.Value, len(e.inputs))
	for i, input := range e.inputs {
		var data []int64
		switch input.kind {
		case miniLMInputIDs:
			data = ids
		case miniLMAttentionMask:
			data = mask
		case miniLMTokenTypeIDs:
			data = tokenTypes
		default:
			return nil, false, fmt.Errorf("unknown MiniLM input kind %d", input.kind)
		}
		tensor, err := ort.NewTensor[int64](shape, data)
		if err != nil {
			destroyValues(inputValues)
			return nil, false, err
		}
		inputValues[i] = tensor
	}
	defer destroyValues(inputValues)

	outputs := []ort.Value{nil}
	e.mu.Lock()
	err := e.session.Run(inputValues, outputs)
	e.mu.Unlock()
	if err != nil {
		return nil, false, err
	}
	if outputs[0] == nil {
		return nil, false, errors.New("MiniLM returned no token embeddings")
	}
	defer outputs[0].Destroy()
	out, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, false, fmt.Errorf("MiniLM output has unexpected type %T", outputs[0])
	}
	vec, err := meanPoolMiniLM(out, mask)
	if err != nil {
		return nil, false, err
	}
	normalizeVector(vec)
	return vec, true, nil
}

func meanPoolMiniLM(out *ort.Tensor[float32], mask []int64) ([]float32, error) {
	shape := out.GetShape()
	data := out.GetData()
	if len(shape) != 3 {
		return nil, fmt.Errorf("MiniLM output rank = %d, want 3", len(shape))
	}
	if shape[0] != 1 {
		return nil, fmt.Errorf("MiniLM batch = %d, want 1", shape[0])
	}
	seqLen := int(shape[1])
	dims := int(shape[2])
	if dims != semanticEmbeddingDims {
		return nil, fmt.Errorf("MiniLM dims = %d, want %d", dims, semanticEmbeddingDims)
	}
	if len(data) < seqLen*dims {
		return nil, fmt.Errorf("MiniLM output too short: %d < %d", len(data), seqLen*dims)
	}
	if len(mask) < seqLen {
		seqLen = len(mask)
	}
	vec := make([]float32, semanticEmbeddingDims)
	var tokens float32
	for tok := 0; tok < seqLen; tok++ {
		if mask[tok] == 0 {
			continue
		}
		tokens++
		base := tok * dims
		for dim := 0; dim < dims; dim++ {
			vec[dim] += data[base+dim]
		}
	}
	if tokens == 0 {
		return nil, errors.New("MiniLM attention mask selected no tokens")
	}
	scale := float32(1) / tokens
	for i := range vec {
		vec[i] *= scale
	}
	return vec, nil
}

func destroyValues(values []ort.Value) {
	for _, value := range values {
		if value != nil {
			_ = value.Destroy()
		}
	}
}
