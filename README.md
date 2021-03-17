# Go Blueprints

From https://www.packtpub.com/product/go-programming-blueprints-second-edition/9781786468949

## Running Containers

```sh
docker-compose up
# Get data on the box. This is really hacky, and should be part of the image.
# Would be nice if docker compose offered an on-run command
docker-compose exec mongo bash
## TODO Make this a sync.Once
export POLL='{"title": "Test Poll", "options": ["jimmycarter", "roygoode", "richardnixon", "arnoldschwarzenegger", "berniesanders"]}'
echo $POLL | mongoimport --db=ballots --collection=polls
```
