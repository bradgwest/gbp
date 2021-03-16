# Go Blueprints

From https://www.packtpub.com/product/go-programming-blueprints-second-edition/9781786468949

## Running Containers

```sh
docker-compose up
# Get data on the box. This is really hacky
docker-compose exec mongo bash
export POLL='{"title": "Test Poll", "options": ["jimmycarter", "roygoode", "richardnixon", "arnoldschwarzenegger", "berniesanders"]}'
echo $POLL | mongoimport --collection=ballots
```
