package numbering

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lagz0ne/c3-design/cli/internal/frontmatter"
	"github.com/lagz0ne/c3-design/cli/internal/walker"
)

// NextContainerId returns the next available container number (1, 2, 3, ...).
func NextContainerId(graph *walker.C3Graph) int {
	containers := graph.ByType(frontmatter.DocContainer)
	if len(containers) == 0 {
		return 1
	}

	max := 0
	for _, c := range containers {
		numStr := strings.TrimPrefix(c.ID, "c3-")
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1
}

// NextComponentId returns the next available component ID for a container.
// Foundation components use slots 01-09, feature components use 10+.
func NextComponentId(graph *walker.C3Graph, containerNum int, feature bool) (string, error) {
	prefix := fmt.Sprintf("c3-%d", containerNum)
	components := graph.ByType(frontmatter.DocComponent)

	var nums []int
	for _, c := range components {
		if !strings.HasPrefix(c.ID, prefix) {
			continue
		}
		numStr := strings.TrimPrefix(c.ID, prefix)
		n, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}

	if feature {
		max := 9
		for _, n := range nums {
			if n >= 10 && n > max {
				max = n
			}
		}
		next := max + 1
		return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
	}

	// Foundation: 01-09
	max := 0
	for _, n := range nums {
		if n >= 1 && n <= 9 && n > max {
			max = n
		}
	}
	next := max + 1
	if next > 9 {
		return "", fmt.Errorf("container c3-%d has no more foundation slots (01-09 full)", containerNum)
	}
	return fmt.Sprintf("c3-%d%02d", containerNum, next), nil
}

// NextAdrId generates an ADR ID with today's date and the given slug.
func NextAdrId(slug string) string {
	date := time.Now().Format("20060102")
	return fmt.Sprintf("adr-%s-%s", date, slug)
}
