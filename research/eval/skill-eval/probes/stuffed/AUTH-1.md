# AUTH-1 Stuffed Probe

## Evidence Commands

- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 search "How is authentication handled and what governs it?" # c3 search
- C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 read c3-213 --full

## Answer

component ref recipe adr contract governance goal/choice/why

Required strings:

- c3 search
- ref-authentication because
- ref-rbac because
- ref-nats-jwt-auth because
- Google OAuth
- test token
- cookie
- UserActor
- currentUserTag
- JWT resolver

IDs:

- recipe-auth-and-access
- c3-213
- c3-202
- c3-209
- ref-authentication
- ref-rbac
- ref-nats-jwt-auth
