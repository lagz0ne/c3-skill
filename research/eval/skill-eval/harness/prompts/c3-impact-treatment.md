## C3 Impact Treatment

For every architecture, blast-radius, or change-impact task, use the frozen local
C3 workflow before ordinary source search. Invoke exactly once:

`bash /opt/c3/bin/c3-impact-bootstrap "<behavior or domain>"`

Do not read `SKILL.md` separately or use bare/global C3. The bootstrap verifies
the frozen skill and sweep sources, routes the behavior into a diverse evidence
pack, inspects forward and reverse relationships, reads bounded evidence, and
returns the closure lanes:
owner/mutation, consumers/state, persistence/event/retry, failure/isolation, and
tests. Classify each as affected, unaffected, or unknown. Pack class and state
describe the stored record, not runtime implementation. Use source anchors to
establish current truth before proposing the change.

Use the remaining calls to source-close every requested lane, starting with the
highest-risk gaps. A pack with zero matching record claims is a low-coverage
signal: search the repository directly for the requested mechanism instead of
expanding adjacent C3 concepts. Before proposing design, give each requested
surface one exact current-mechanism anchor or mark it unknown with the next
source check. Use no more than ten additional tool calls after the bootstrap;
batch related surfaces in one search or read. Keep each source result below 40
lines and roughly 1 KB.

A treatment answer is invalid unless route, graph, and evidence commands
succeed. Distinguish direct from transitive impact, explain the causal and
failure boundary, and propose a safe pre-change scope. Unsupported impact stays
unknown with the next check named.
