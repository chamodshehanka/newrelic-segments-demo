# New Relic Segments Demo

This repository contains a demo application that showcases the integration of New Relic Segments for performance monitoring and analytics. The demo application is built using [your chosen technology stack, e.g., Node.js, React, etc.] and demonstrates how to instrument your code with New Relic Segments to gain insights into application performance.

## Curls

```shell
# Untraced flow:
curl -i --max-time 20 http://localhost:8080/process-untraced

# Traced flow (distributed trace headers are injected by the agent in chamod):
curl -i --max-time 30 http://localhost:8080/process-traced

```