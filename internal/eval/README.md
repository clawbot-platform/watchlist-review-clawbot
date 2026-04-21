# Eval expansion notes

This folder gains a second deterministic suite:

- `deterministic-history-relationship-v1.json`

Recommended execution order:

1. run current calibrated suite
2. run history/relationship suite
3. fail promotion if baseline regresses
4. optionally gate on new suite once scoring hooks are live
