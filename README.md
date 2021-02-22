# Go Blueprints

From https://www.packtpub.com/product/go-programming-blueprints-second-edition/9781786468949

## Running Containers

```sh
docker run --name mongod -v $(pwd)/db/mongo:/data/db -p 27017:27017 mongo
docker run --name nsqd -p 4150:4150 -p 4151:4151 \
    nsqio/nsq /nsqd \
    --broadcast-address=172.17.0.3 \
    --lookupd-tcp-address=172.17.0.3:4160
docker run --name lookupd -p 4160:4160 -p 4161:4161 nsqio/nsq /nsqlookupd
```
