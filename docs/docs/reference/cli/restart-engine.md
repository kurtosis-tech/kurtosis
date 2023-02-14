---
title: Restart the engine
sidebar_label: Restart the engine
slug: /restart-engine
sidebar_position: 2
---

### Restart the engine
The CLI interacts with the Kurtosis engine, which is a very lightweight container. The CLI will start the engine container automatically for you and you should never need to start it manually, but you might need to restart the engine after a CLI upgrade. To do so, run:

```bash
kurtosis engine restart
```