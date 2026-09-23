# GoAkt Actor Example — Gate State Machine (Become / UnBecome)

This example demonstrates how to build a simple **state machine actor** using  
**GoAkt v4** and the `Become` / `UnBecome` behavior‑switching mechanism.

The actor (`Gate`) has two modes:

- **Closed** → cannot be used  
- **Open** → can be used  

The actor switches between these modes dynamically based on incoming messages.

---

## 🚦 Concept

The Gate actor supports three message types:

| Message | Meaning |
|---------|---------|
| `Open`  | Switch gate to **open** mode |
| `Close` | Switch gate back to **closed** mode |
| `Use`   | Attempt to use the gate |

The actor starts in **closed** mode.

---

## 🧠 Behavior Switching

GoAkt allows actors to change their message handler at runtime:

- `ctx.Become(newBehavior)`  
  Switches to a new behavior function.

- `ctx.UnBecome()`  
  Restores the previous behavior.

This makes it easy to model **finite state machines**, workflows, and session logic.

---

## 🏗️ Actor Implementation

The `Gate` actor defines two behaviors:

### Closed Behavior

```go
func (g *Gate) closedBehavior(ctx *goakt.ReceiveContext) {
    switch ctx.Message().(type) {
    case *Open:
        ctx.Become(g.openBehavior)
        ctx.Response("gate opened")
    case *Use:
        ctx.Response("cannot use: gate is closed")
    default:
        ctx.Unhandled()
    }
}
