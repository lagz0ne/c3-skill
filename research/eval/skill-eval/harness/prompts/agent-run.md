# Agent Run Prompt

You are evaluating the C3 skill on a real architecture growth task. Use only
the C3 instructions and references supplied in this packet. If your environment
supports skills, treat the supplied `skills/c3/SKILL.md` as the governing C3
skill; do not use any installed or global C3 skill.

Work inside the isolated project at `/work/project`.

Use the local C3 wrapper only:

```bash
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh <command>
```

Do not use bare `c3x`. Do not load C3 from home-directory skills or marketplace
installs.

Your output should prove whether the skill helps with grow-as-you-go C3 work.
Create or update project C3 docs as needed in the isolated project, then return
a compact final report with these sections:

- `Evidence commands`: exact C3 commands run in order.
- `Architecture growth`: containers, components, concepts, and boundaries added.
- `Feature growth`: changes that increase product complexity.
- `Migration and gardening`: how old docs were raised to the new rung without
  leaving mixed contracts.
- `Verification`: checks run and current result.
- `Caveats`: anything not proven.

Hard expectations:

- Model the system as multiple containers, including frontend, backend,
  integration, and database.
- Document concepts and boundaries across containers and components.
- Use ADR/change-unit or migration-style C3 flows for growth of existing facts
  when the project has existing C3 docs.
- Treat canvas growth as rung growth: facts are complete to their current rung;
  when the system earns richer sections, raise the contract and migrate affected
  facts completely.
- Include migrations and document gardening as first-class work.
- Avoid codemap work unless absolutely necessary for the growth task. This eval
  is about documentation/system growth, not code-file mapping.
- Keep the answer compact enough for small/free models, but make it concrete.
