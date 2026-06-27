# Development Principles

## TDD by Kent Beck

**Three Laws:**
1. Write no production code until you have a failing test.
2. Write only enough of a test to make it fail (compilation failure counts).
3. Write only enough production code to make the failing test pass.

**Red-Green-Refactor cycle:**
- **Red** — write a small test that fails. One test, one concept.
- **Green** — write the simplest code that makes it pass. Don't optimize yet.
- **Refactor** — clean up duplication and design, keeping all tests green.

**Rules:**
- One test at a time. Never write the next test until the current one passes.
- Tests are documentation. They describe the intended behavior precisely.
- If a test is hard to write, the design is telling you something.

## YAGNI — You Aren't Gonna Need It

Don't implement something until it is actually needed.
Speculative features add complexity without delivering value.
Delete code you don't use. The best code is no code.

## KISS — Keep It Simple, Stupid

Prefer the simplest solution that solves the actual problem.
Complexity is a liability, not an asset.
If you can't explain what a piece of code does in one sentence, simplify it.

## Lean — Eliminate Waste

**Seven wastes (applied to software):**
1. **Overproduction** — building features nobody asked for.
2. **Waiting** — blocked PRs, slow CI, unreviewed code.
3. **Defects** — bugs that reach production and require rework.
4. **Overprocessing** — more ceremony, abstraction, or process than the problem requires.
5. **Inventory** — unmerged branches, unshipped features, stale drafts.
6. **Motion** — context switching, unclear requirements, missing tooling.
7. **Transportation** — handing off work without shared understanding.

**Flow over batch:** small incremental changes ship value faster and fail safer than large batches.
**Amplify learning:** feedback loops (tests, CI, code review) should be as short as possible.
**Decide as late as responsible:** delay irreversible decisions until you have enough information.
