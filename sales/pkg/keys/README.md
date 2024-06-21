# The Interesting Face of Identifiers
Using serial/incremental IDs can offer some benefits:

1. Simplicity/Readability: Serial IDs are straightforward to implement and understand. They increment sequentially, making them easy to generate and manage.
2. Performance: In databases, serial IDs can improve indexing and query performance. Indexes on sequential IDs tend to be more efficient due to their ordered nature.
3. Efficiency: Generating a new serial ID is computationally inexpensive compared to other methods like UUIDs.
4. Storage: Serial IDs require less storage space compared to UUIDs, which can be beneficial in larger-scale applications.

However, there are known drawbacks to using serial IDs, such as predictability and the risk of enumeration attacks. Where security and obfuscation are considered, other ID generation methods like UUIDs are typically reached for and conversations around this topic typically finish here. But even in this case, recurring use of the same ID for the same entities encourages side-channel attacks; traffic analysis or some form of fingerprinting that can link interaction to users.

## An overkill approach to internal vs. external IDs
This package, designed as a thought experiment in this arena (ie, should I actually be doing this, why is this not necessary etc...) - allows for all the benefits of serial/incremental ID's, creating a mapping between internal and an issuing of external ID's.

### KeySet
A `keyset` is responsible for the verifying/decoding/issuing of external ID's. We need;
- A set of `RSA` `PEM` formatted public/private keys.
- An AES encryption key (we have chosen 16 byte size, hence operating on 128 bit blocks).
- An expiry of sorts and internally we will allow for `keyset` revoking.
- An associated UUID for universal address.

#### Rotation
At any time, we can rotate the `keyset` by marking a key revoked. Also tune frequency of rotation `keyset`s via shorter `expiry` registered `keyset`.

Request middleware handles stale identifier references.

### Codec
The external ID's are `RSA256` (asymmetric) signed JWT tokens, that have public claims `exp`, `kid` and `internal_id`.

- `kid` is a UUID of the `keyset` used to encode this external ID.
- `exp` is a unix timestamp of when this external ID will expire.
- `internal_id` is a `AES-GCM` symmetric encrypted identifier (benchmark for this encryption process lives in `encrypt_test.go`).

### Defensive Advantages
When a resource mutation request is submitted, an issued external ID will be provided. This serves as an instrument of threat identification as we now have the following visibility of various request attack vectors;

- Enumeration attacks will result in large number of requests of invalid external ID's being passed in (verifiable by JWT signature).
- Due to the one **to many to one** symmetric encryption scheme;
    - Single valid tokens used unusually many times can be easily identified.
        - We can also provide short TTL blacklist caches of external ID's and soft-enforce "use-once" external ID's
    - No two encryptions of the same ID's will be the same, hence pattern identification is also mitigated.
- We can have various `keyset`'s active at the same time.
- We can easily revoke any compromised `keyset`'s.

### TradeOff
With the various defensive advantages we deal with various performance tradeoff's;
- Added database lookup for `kid`'s
    - We reduce this by holding a pull-through cache of a `keyset` "chain" with regular polling of "active" keys in that "chain".
- External ID's are now complex to create
    - We monitor this with benchmarks (`encrypt_test.go` for the `ACM-128-GCM` scheme...)

## Integration
We introduce an `OpaqueID` type that satisfies marshaller interfaces of various integrations. An `OpaqueID` can be loaded with any internal or external ID, but does require a registration with a `codec` in order to perform the back and fourth mapping between the two.
