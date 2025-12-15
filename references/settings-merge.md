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

## When to Customize

### Decision Matrix

| Scenario | Use Defaults | Override | Disable |
|----------|--------------|----------|---------|
| Standard project | Yes | No | No |
| Strict include/exclude rules | Yes | Add to lists | No |
| Custom litmus test | Yes | Replace | No |
| Different diagram tool | Yes | Set tool | No |
| No diagrams at layer | Yes | No | Yes (`useDefaults: false`) |
| Team-specific guardrails | Yes | Add guard section | No |

### Common Customization Scenarios

**Scenario 1: Add custom include/exclude rules**
```yaml
container:
  useDefaults: true  # Keep defaults
  include:
    - "Custom element for this project"
  exclude:
    - "Element we never document"
```

**Scenario 2: Replace litmus test**
```yaml
component:
  litmus: "Would a new team member know how to modify this?"
```

**Scenario 3: Disable defaults entirely**
```yaml
context:
  useDefaults: false  # Only use explicit settings
  include:
    - "System boundary"
    - "Actors"
  # exclude, litmus, diagrams all from settings only
```

### Anti-Patterns

| Don't | Why | Do Instead |
|-------|-----|------------|
| Override everything | Defeats purpose of shared defaults | Override only what's different |
| Empty useDefaults: false | No guidance at all | Keep defaults or explicitly list all |
| Conflicting include/exclude | Confusing behavior | Keep lists disjoint |
| Custom litmus without understanding default | May miss important criteria | Understand default before replacing |
