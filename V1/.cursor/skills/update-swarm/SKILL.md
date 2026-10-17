---
name: update-swarm
description: >-
  Runs the MiniVerse update swarm before and after a product change.
  Use when the user asks to add, change, design, or ship an update to
  MiniVerse, including a new simulation behavior, API, page behavior,
  or product idea. Also use when the user mentions the swarm, project
  manager, physicist, chemist, frontend, backend, or tester role.
---

# MiniVerse update swarm

You orchestrate. Do not skip the gates below and start coding.

The roles are the project agents in `.cursor/agents/`. Delegate to each one with the Task tool and that agent's name. Pass the user request, the product-lead verdict, and any constraints already returned. Reviewers return a verdict and do not edit files. Implementers edit only their own tree.

## When to run

Run this for an update: a new behavior, a change to existing behavior, or a product idea the user wants built.

Do not run it for a question, a review of existing code, or a typo in a comment.

## Who is called

Every update starts with `product-lead` and ends with `tester`.

| The update touches | Also call |
| --- | --- |
| Heat, temperature, distance, time, phase, mass, or whether a suggestion is physically plausible | `physicist` |
| Species, reactions, atoms, the catalog, or a new material | `chemist` |
| `chem/`, `cmd/`, the HTTP API, or the frame JSON | `backend-architect` |
| `web/`, the canvas, or view state | `frontend-developer` |

Call `physicist` and `chemist` on a suggestion before anyone builds it, when the suggestion is about the world. They add common sense. They do not implement.

## Order

1. `product-lead`. If `Proceed` is no, stop and tell the user. Do not replace the user's request with a different idea.
2. `physicist` and `chemist` together, when either applies. A veto stops the build. Report it and wait for the user.
3. `backend-architect`, when the server or simulation changes. It states the frame contract before the client is written.
4. `frontend-developer`, when the page changes. Start it in parallel with the backend only when the frame JSON does not change.
5. `tester`. On a critical failure, send that failure back to the owner of that tree, then run `tester` again.
6. `product-lead` once more, with what shipped, to check the benefit still holds. This pass adds no scope.

## What you tell the user

Lead with the user benefit, then what changed, the constraints that bound it, and the test result. Leave each agent's full writeup out unless the user asks for it.
