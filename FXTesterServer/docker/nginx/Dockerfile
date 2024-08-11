FROM nginx

COPY ./docker/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY ./settings/cert/fx-tester.key.pem  /etc/nginx/conf.d/key.pem
COPY ./settings/cert/fx-tester.cert.pem /etc/nginx/conf.d/cert.pem