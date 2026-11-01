# Context-aware Enchanted validation

## Decision

Enchanted release validation treats the server-sent Hub form as runtime data instead of hard-coding one child name or one hop count. The agent records the current form entries, classifies available children, maintenance or status entries, navigation entries, and informational entries, then tests every actionable top-level child from a fresh Hub session.

When a selection sends another form, the agent treats it as a nested menu and selects one available child from that menu. When a selection sends `Transfer`, the agent follows every observed hop, completes every offered resource-pack exchange, reaches the terminal spawn, and retains the normal stability interval. The agent returns to the Hub before testing the next top-level child.

A selection that does not send `Transfer` is not automatically a failure. The agent must determine whether the selection produced a direct playable route, a maintenance or unavailable response, a navigation action, an informational response, or an unchanged Hub session. Transfer and child-resource-pack checks are conditional: reports use `pass` when the observed flow succeeds and `not_applicable` only when the reviewed route matrix proves that the current flow did not use that operation.

The sanitized Enchanted report requires `enchanted_form_review=pass` and `enchanted_route_matrix=pass`. At least one currently playable route must reach spawn and stability. The route ledger remains review evidence and must not contain raw form payloads, credentials, packet data, resource-pack keys, or other sensitive identifiers.

## Rationale

Enchanted changes its form entries and route topology. Some entries can be unavailable, some can open another menu, and some can remain on the current Hub without a `Transfer`. Requiring every release session to select one historical child and observe one mandatory hop can turn valid server behavior into a false release failure. Context-aware classification preserves the technical gates for every observed transfer and pack exchange while allowing the release evidence to follow the server's current behavior.

This decision supersedes the fixed-child wording in earlier Enchanted validation notes. It does not weaken the general session gates: upstream authentication, offered-pack completion and decryption, spawn, bidirectional traffic, stability, capture integrity, and absence of unexplained failures remain required.

## Evidence boundary

The protocol client observes the server's current form and routes. It does not claim retail rendering or gameplay. A route marked maintenance, unavailable, or not applicable must retain enough local evidence for the agent to explain that classification. A route with an unexplained timeout, disconnect, decode error, or failed pack exchange remains a validation failure and starts the normal fix-forward investigation.
