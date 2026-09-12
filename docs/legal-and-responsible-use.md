# Resource-pack legal and responsible-use assessment

Reviewed on 2026-09-13. This is a technical and project-risk assessment, not legal advice. The statutory discussion is primarily about United States law because "DMCA" refers to United States law. Contract, copyright, computer-access, and reverse-engineering rules vary by jurisdiction and by the terms that apply to a particular account, server, pack, and user.

## What the implementation actually does

The following trace is based on the current code and gophertunnel `v1.61.0`, not on the names of functions or on an assumed client flow.

1. A downstream Bedrock client connects to the local proxy. During its login sequence, [`internal/proxy/proxy.go`](../internal/proxy/proxy.go) uses that client's `ClientData` to establish one upstream connection to the configured server.
2. Upstream Xbox device authentication is the default. A successful device login is cached per user and reused until refresh fails, but every upstream connection still uses the resulting authenticated session. `--auth none` instead sends an offline identity and succeeds only if the selected server accepts it. A successful connection is therefore a server-accepted current session, but that protocol fact is not by itself a legal conclusion about copyright permission.
3. The upstream server sends `ResourcePacksInfo` after the Bedrock login handshake. Each `TexturePackInfo` carries the pack UUID, version, compressed size, `ContentKey`, content identity, flags, and possible download URL. BedrockDebugProxy does not look up or derive the master key. gophertunnel copies the exact server-supplied `ContentKey` into its current connection's pack queue.
4. BedrockDebugProxy currently returns `true` for every upstream pack offer. This is an automatic protocol acceptance, including for a pack that a graphical client might have allowed a person to decline. For an offer with an HTTP(S) `DownloadURL`, the proxy fetches only that URL during the current upstream session, validates the archive against the advertised UUID, version, and size, and supplies it through gophertunnel's per-connection cache. Packs without a usable URL, or failed URL downloads, use the established RakNet chunk path as a visible fallback.
5. The same downloaded `resource.Pack` objects are returned by the downstream listener's `FetchResourcePacks` callback. gophertunnel sends the downstream client a new `ResourcePacksInfo` containing the same content key and then services the client's pack requests. A normal client on this proxy path therefore also receives the archive and key through the protocol.
6. [`internal/bedrock/resource_pack.go`](../internal/bedrock/resource_pack.go) first records the exact reconstructed encrypted archive. Its event metadata includes the content key, reconstructed-pack URL if present, manifest, identifiers, endpoints, flags, size, and checksum. For a successfully fetched URL pack, the URL is retained on the reconstructed pack; if URL retrieval fails, the raw `ResourcePacksInfo` packet event remains the authoritative record of the advertisement while any raw pack-chunk fallback events independently retain their transfer payloads.
7. Only when `--decrypt-resource-packs` is set, the recorder passes that current pack and its attached server-supplied key to [`internal/resourcepack/decrypt.go`](../internal/resourcepack/decrypt.go). There is no key search, guessing, brute force, credential bypass, server exploit, or alternate key-recovery path. Invalid key lengths and undocumented variants are rejected.
8. Decryption is not memory-only. Each `contents.json` ciphertext and plaintext is handled in memory within documented limits. Per-file content is decrypted as a stream into a mode-restricted ZIP in the operating system's temporary directory. That file is removed on normal success or failure, but an abnormal process or machine termination can leave residue for local cleanup. Plaintext persists by design inside the capture as content-addressed `resource_pack.contents_manifest` and `resource_pack.decrypted_archive` blobs. The exact encrypted archive remains alongside them.
9. The shipped CLI has no offline resource-pack decrypt command and no parameter for an arbitrary pack or caller-supplied key. The format decoder is an internal Go package and is generic enough for tests and the live integration to pass a `ReaderAt` and key. A developer could modify the source to expose it, which is why that expansion requires explicit review.
10. The project has no upload, publishing, or redistribution path. `export` makes a complete local `.bdpcap` copy, including raw packets, master and per-file keys, and any decrypted assets. A user can manually move or publish local files, but that is not performed by BedrockDebugProxy.

## Separate legal and policy questions

### Copyright infringement

Copyright owners have separate reproduction and distribution rights under [17 U.S.C. 106](https://www.copyright.gov/title17/92chap1.html). A local plaintext copy and a public redistribution are different uses. Research is one purpose named in [17 U.S.C. 107](https://www.copyright.gov/title17/92chap1.html), but fair use remains a case-specific balance of purpose, the creative nature of the work, amount used, and market effect. "Debugging" or "noncommercial" does not automatically make copying fair use. Persisting an entire creative pack weighs differently from retaining only protocol facts or small necessary excerpts. Keeping a copy private and avoiding market substitution reduce some practical and fair-use concerns, but do not create a blanket private-use exception.

Tool source that contains no pack assets presents a different copyright question from a repository containing textures, models, sounds, scripts, or a decrypted archive. A rights holder could still assert an anti-circumvention claim against the tool, but that is not the same claim as direct infringement by copying the pack into the repository.

### Anti-circumvention

[17 U.S.C. 1201](https://www.copyright.gov/title17/92chap12.html) separately addresses the act of circumventing an effective access control and trafficking in certain circumvention technology. The statute defines circumvention to include decrypting an encrypted work "without the authority of the copyright owner." It also defines an effective access control by whether ordinary operation requires information or a process applied with that authority.

The strongest implementation-specific fact reducing risk is that the server itself transmits the exact key in the normal `ResourcePacksInfo` path to the current client session, and the proxy applies that key without defeating server authentication or recovering a missing secret. A plausible authorization argument is that applying information deliberately supplied for this client is the ordinary authorized operation rather than unauthorized avoidance. The difficult unresolved question is whether authority to let the game client use the pack also authorizes a debugging proxy to persist a general plaintext ZIP. BedrockDebugProxy must not claim that server delivery conclusively answers that legal question.

Publication of source also raises a distinct anti-trafficking question. The project as a whole has extensive non-decryption uses, decryption is opt-in, no keys are bundled, and the tool is not marketed as a way to obtain protected assets. Those facts reduce risk under statutory tests concerning primary design, other significant uses, and marketing. They do not guarantee that a claimant or court will agree, particularly when looking at the decryption component by itself.

United States appellate approaches are not uniform. The Federal Circuit's [Chamberlain v. Skylink](https://law.justia.com/cases/federal/appellate-courts/F3/381/1178/608254/) analysis emphasized authorization and a relationship to infringement. The Ninth Circuit's official [MDY Industries v. Blizzard opinion](https://cdn.ca9.uscourts.gov/datastore/opinions/2011/02/17/09-15932.pdf) rejected a general infringement-nexus requirement for access-control claims and found liability involving a World of Warcraft access control. These decisions concern different facts and jurisdictions, but they are a warning against presenting one authorization theory as settled nationwide law.

### Access control and computer access

Current code opens one ordinary Bedrock connection and asks for the packs the server advertised on that connection. It does not scan another account, query a hidden pack endpoint, reuse a key from another server, evade a ban, forge ownership, or continue when the server rejects login. This materially separates the current feature from unauthorized remote access.

That boundary does not grant permission from a server owner. `--auth none` should be used only where the server permits offline login, and testing a server without its authorization can raise contract, computer-access, privacy, or abuse issues independent of copyright. The statutory security-testing provision in [17 U.S.C. 1201(j)](https://www.copyright.gov/title17/92chap12.html) itself requires authorization from the owner or operator of the tested computer, system, or network.

### Reverse engineering, interoperability, and research

The permanent interoperability exception in [17 U.S.C. 1201(f)](https://www.copyright.gov/title17/92chap12.html) is narrow. It concerns lawfully used computer programs, elements necessary for interoperability of an independently created program, noninfringing acts, and tightly limited sharing of information or means for that purpose. Resource packs also contain creative textures, sounds, models, and other material that may not fit a computer-program analysis. The project should describe interoperability and protocol research as its purpose, but should not claim that section 1201(f) automatically covers every pack or every use.

The encryption-research exception in [17 U.S.C. 1201(g)](https://www.copyright.gov/title17/92chap12.html) requires good-faith research into flaws or vulnerabilities in encryption technology, lawful acquisition, necessity, an effort to obtain authorization, noninfringement, and other factors. Applying a known AES mode with a supplied key for debugging is not necessarily research into an encryption flaw. Do not rely on this exception as the general rationale for the feature.

The current temporary exemption for [good-faith security research in 37 C.F.R. 201.40(b)(18)](https://www.copyright.gov/title37/201/37cfr201-40.html) is also limited to security flaws or vulnerabilities, authorized or lawfully acquired systems, harm avoidance, and security-promoting use that does not facilitate infringement. The regulation explicitly says it is not a safe harbor from other laws. The triennial process waives only specified acts for specified classes and purposes. It is not blanket permission for tooling or redistribution.

### Contract and platform policy

The current [Minecraft EULA](https://www.minecraft.net/en-us/eula) allows independent tools in its summary when they do not appear official, but also restricts distribution of Minecraft software and content, including downloadable and Marketplace content. The [Minecraft Usage Guidelines](https://www.minecraft.net/en-us/usage-guidelines) define assets to include code, textures, models, sounds, and other game material, prohibit redistribution of game files, require independent projects not to appear official, and make clear that their permissions can change. They do not specifically approve this decryption workflow.

The current [Microsoft Services Agreement](https://www.microsoft.com/en-us/servicesagreement) separately restricts bypassing technological protection measures and decrypting or reverse engineering software or other aspects of the services except where applicable copyright law expressly permits it. A Featured Experience, Creator Experience, Marketplace pack, or independent server may also have different owners and additional terms. Contract and account-enforcement risk can exist even where a statutory defense might apply.

GitHub's [DMCA Takedown Policy](https://docs.github.com/en/site-policy/content-removal-policies/dmca-takedown-policy) treats ordinary infringement notices and technical circumvention claims differently. GitHub asks a circumvention claimant to identify the measure, how it controls access, and how the project circumvents it, then says technical and legal experts review the claim and normally contact the developer before disabling the repository. That process reduces the chance of an unexplained instant circumvention takedown, but it does not prevent complaints, ordinary notices about bundled assets, temporary disruption, or litigation. Section 512 notice and counter-notice procedure is a hosting safe-harbor process, not a ruling that the tool or content is lawful.

## Scenario risk comparison

These ratings are project-risk judgments, not legal outcomes.

| Scenario | Relative risk | Main reasons |
| --- | --- | --- |
| 1. Public source, no copyrighted packs | Lower for direct infringement, moderate and uncertain for anti-circumvention and contract | No third-party asset is copied into GitHub. The source still includes a decryption component, so design, marketing, authority, exemptions, and applicable terms remain relevant. |
| 2. Decrypt a pack and key sent to this current server-accepted session | Moderate and fact-dependent | Normal protocol delivery, the actual server-supplied key, no cracking, and no auth bypass reduce risk. Persisting plaintext can exceed what a rights holder says was authorized for gameplay. Default Xbox authentication helps establish session provenance but does not establish copyright permission. |
| 3. Save plaintext locally for debugging | Moderate | Private research purpose, controlled storage, and no market substitution reduce practical exposure. A full persistent copy still implicates reproduction, contract, confidentiality, and possible section 1201 questions. Fair use and private use are not automatic defenses. |
| 4. User redistributes a decrypted Featured or Creator Experience pack | High to very high | Whole creative assets can substitute for authorized access, affect a market, violate terms, expose confidential material, and produce a straightforward copyright claim by Microsoft, a partner, a server operator, or another creator. The source-code license grants no redistribution right for the pack. |
| 5. Repository accidentally contains copyrighted assets or decrypted packs | Very high and avoidable | This creates direct hosted copies and a much simpler notice target than source-only tooling. Git history and forks can prolong exposure even after deletion. Keys and private URLs may add security or confidentiality claims. |
| 6. Add a generic offline command for arbitrary pack plus caller-supplied key | Higher than the current live-only CLI | It severs session provenance, broadens the intended audience and likely uses, makes misuse easier, and weakens the explanation that artifacts were delivered to the current client. It could strengthen anti-trafficking and contract arguments even though the user supplies the key. This capability is not currently exposed and requires explicit maintainer review before implementation. |
| 7. Find, crack, brute-force, recover, or bypass a key not supplied to the accepted client session | Highest | This removes the strongest authorization and provenance facts and materially increases anti-circumvention, unauthorized-access, contract, security, privacy, and abuse risk. It is outside project scope and must not be added as a routine feature. |

## Factors that change the risk

Risk increases when a pack is complete and highly creative, Marketplace or commercially licensed content is involved, plaintext is retained indefinitely, keys or private URLs are exposed, use is commercial, a user lacks server or account authorization, the tool is marketed for asset extraction or DRM bypass, content is uploaded, or a feature accepts packs and keys unrelated to the current session.

Risk decreases, without disappearing, when testing uses an owner-controlled server and synthetic pack, the account and server session are legitimate, the exact key provenance is retained, decryption is opt-in, raw evidence remains authoritative, output stays local and access-controlled, only necessary research results are published, third-party assets are not committed, and the project is described as an independent debugging and interoperability tool rather than an extraction utility.

Important unknowns include the copyright owner and license of each file, server-specific terms, whether the rights holder considers local plaintext storage authorized, whether the pack is a computer program or primarily creative media for a particular exception, the user's jurisdiction, which United States circuit would apply, and the actual purpose and dissemination of a specific capture. Counsel familiar with copyright, section 1201, software interoperability, and the relevant jurisdiction should review the public-release posture and any real complaint.

## Project controls and non-goals

- Keep decryption opt-in and preserve the exact encrypted archive first.
- Keep the production key path tied to the current upstream connection's server-supplied `ContentKey`.
- Keep decrypted artifacts local, sensitive, and complete in capture exports. Do not add automatic upload or publication.
- Do not add an offline decrypt CLI, remote pack crawler, key database, key recovery, brute force, access-control bypass, or redistribution workflow without explicit technical and legal review.
- Use synthetic fixtures in tests and documentation. Do not commit real packs, captures, keys, private URLs, or extracted assets.
- When publishing research, prefer protocol facts, hashes, field layouts, minimal necessary excerpts, and reproducible synthetic vectors over copyrighted assets.
- If a rights holder sends a complaint, preserve the notice and exact repository revision, separate an asset claim from a circumvention claim, do not make admissions or submit a counter-notice reflexively, and obtain qualified legal advice.

## Licensing and ownership

`GPL-3.0-or-later` covers BedrockDebugProxy source and covered derivatives. It does not transfer ownership of data merely processed by the program. GPLv3 itself states that output is covered only when its content constitutes a covered work. A captured or decrypted pack, texture, model, sound, script, server URL, identifier, or player communication keeps whatever copyright, license, confidentiality, privacy, and contractual status it had before capture.

Dependencies remain under their own licenses as recorded in [`THIRD_PARTY_NOTICES.md`](../THIRD_PARTY_NOTICES.md). A `.bdpcap` can contain all three categories at once: BedrockDebugProxy metadata, protocol observations, and third-party content. The project license does not grant permission to redistribute the third-party portions. The operator is responsible for determining authority to collect, retain, disclose, and redistribute a capture.

## Claims the project must not make

Do not describe this feature with any of the following claims.

- "The server sent the key, so every use is authorized or legal."
- "Research, interoperability, security testing, fair use, or a DMCA exemption automatically protects the project."
- "Anything a normal client can receive may be freely extracted or redistributed."
- "The feature is memory-only" or "no plaintext is written to disk."
- "The proxy cracks, breaks, bypasses, guesses, or recovers Marketplace keys."
- "The tool bypasses server authentication" or "works without a server-accepted session."
- "All captured or decrypted content becomes GPL."
- "A local copy can never infringe" or "noncommercial use is always fair use."
- "The implementation supports every encrypted Bedrock or Marketplace pack."
- "Decryption success proves plaintext integrity."
- "BedrockDebugProxy is official, approved, endorsed, or associated with Mojang or Microsoft."
