# A Rekor monitor
This project hosts the code and deployment configuration for [rekor-monitor.flxw.de](https://rekor-monitor.flxw.de).
It is my attempt to mine some analytics data on how Rekor, Sigstore's transparency log, is used and who are the main users.
Furthermore, I hope to evolve it into a monitoring system for Rekor, where users can be notified of suspicious signatures made in their name.

## Local Deployment

You can run this project locally in a containerized fashion using the following commands:

```bash
git clone github.com/flxw/rekor-monitor
cd rekor-monitor/local
docker-compose up
```

This will build the Rekor crawler image and container, and start it along with the PostgreSQL and Grafana containers.
## Acknowledgements

 - [Sigstore](https://sigstore.dev)
 - [Rekor](https://github.com/sigstore/rekor)
 