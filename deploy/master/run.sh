cd dating/stg
docker load -i oauth-dating.tar
docker rm -f oauth-dating-prd

docker run -d --name oauth-dating-prd \
  -e GINPORT=3000 \
  -e MDB_MGO_URI="mongodb://oauth:AuidwEyf776GG2S@172.31.23.58:27018/oauth" \
  -p 4001:3000 \
  oauth-dating

exit