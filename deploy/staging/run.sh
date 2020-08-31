cd dating/stg
docker load -i oauth-dating.tar
docker rm -f oauth-dating-stg

docker run -d --name oauth-dating-stg \
  -e GINPORT=3000 \
  -e MDB_MGO_URI="mongodb://oauth:AuidwEyf776GG2S@172.31.27.200:27017/oauth_stg" \
  -p 4000:3000 \
  oauth-dating

exit