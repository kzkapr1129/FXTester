#!/bin/zsh

<< 本スクリプトについて
本スクリプトでは/etc/hostsに下記の内容を書き込む。

※ 注意
 この設定はdocker環境を含む開発段階で必要な設定であり、
 サーバーにIPを個別に割り振るような本番環境では不要な設定。

127.0.0.1 fx-tester-fe
127.0.0.1 fx-tester-be
127.0.0.1 fx-tester-db
127.0.0.1 keycloak
本スクリプトについて

if ! grep -q "fx-tester-fe" "/etc/hosts"; then
    echo "127.0.0.1 fx-tester-fe" >> /etc/hosts
fi

if ! grep -q "fx-tester-be" "/etc/hosts"; then
    echo "127.0.0.1 fx-tester-be" >> /etc/hosts
fi

if ! grep -q "fx-tester-db" "/etc/hosts"; then
    echo "127.0.0.1 fx-tester-db" >> /etc/hosts
fi

if ! grep -q "keycloak" "/etc/hosts"; then
    echo "127.0.0.1 keycloak" >> /etc/hosts
fi