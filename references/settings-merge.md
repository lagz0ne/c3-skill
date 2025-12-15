# Settings Merge Logic

Shared settings loading pattern for all C3 layer skills.

## Usage

Each layer skill should load and merge settings at start:

```bash
# Check for project settings
cat .c3/settings.yaml 2>/dev/null
```

## Merge Algorithm

```xml
<settings_merge layer="[context|container|component]">
  <!-- Step 1: Load defaults from this skill's defaults.md -->
  <defaults source="skills/c3-{layer}-design/defaults.md">
    <!-- Layer-specific defaults loaded here -->
  </defaults>

  <!-- Step 2: Check settings.yaml for layer section -->
  <project_settings source=".c3/settings.yaml">
    <if key="{layer}.useDefaults" value="false">
      <!-- Don't load defaults, use only project settings -->
    </if>
    <if key="{layer}.useDefaults" value="true" OR missing>
      <!-- Merge: project settings extend defaults -->
      <include>defaults + {layer}.include</include>
      <exclude>defaults + {layer}.exclude</exclude>
      <litmus>{layer}.litmus OR default</litmus>
      <diagrams>{layer}.diagrams OR default</diagrams>
      <guidance>{layer}.guidance (layer-specific prose)</guidance>
    </if>
  </project_settings>

  <!-- Step 3: Also load global settings -->
  <global>
    <diagrams_tool>settings.diagrams (e.g., mermaid)</diagrams_tool>
    <guard>settings.guard (team guardrails)</guard>
  </global>
</settings_merge>
```

## Output Format

Display active configuration after merge:

```
{Layer} Layer Configuration:
├── Include: [merged list]
├── Exclude: [merged list]
├── Litmus: [active test]
├── Diagrams: [tool] - [types]
├── Guidance: [layer-specific notes]
└── Guardrails: [if any]
```

## Apply Throughout

Use loaded settings when:
- Deciding what belongs at this layer (litmus test)
- Making diagram decisions (override defaults if settings specify)
- Applying team guardrails
- Writing documentation (guidance from settings)
- Checking include/exclude rules for content placement
