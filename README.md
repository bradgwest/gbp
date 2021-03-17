# Go Blueprints

From https://www.packtpub.com/product/go-programming-blueprints-second-edition/9781786468949

## Running Containers

```sh
docker-compose up
# Get data on the box. This is really hacky, and should be part of the image.
# Would be nice if docker compose offered an on-run command
## TODO Make this a sync.Once
echo '{"title": "Test Poll", "options": ["jimmycarter", "roygoode", "richardnixon", "arnoldschwarzenegger", "berniesanders"]}' | mongoimport --db=ballots --collection=polls
```

## Running Programs

```sh
# from cmd/twittervotes
./twittervotes -mongoAddr 0.0.0.0:27017 -nsqAddr 0.0.0.0:4150
# from cmd/counter
./counter -lookupAddr 0.0.0.0:4161 -mongoAddr 0.0.0.0:27017
```
