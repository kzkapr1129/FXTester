#!/bin/zsh

baseDir="./taskfiles/env/darwin"

[[ -r $baseDir/anyenv/init.sh ]] && sh $baseDir/anyenv/init.sh
[[ -r $baseDir/direnv/init.sh ]] && sh $baseDir/direnv/init.sh

exit 0