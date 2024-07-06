#!/bin/zsh

# anyenvの有効化
if ! grep -q "anyenv init -" ~/.zshrc; then
  if command -v anyenv 2>&1 >/dev/null; then
    # .zshrchへの書き込み
    echo "eval \"\$(anyenv init -)\"" >> ~/.zshrc
    # direnvの有効化(.zshrcの書き込みした直後は.zshrcの再適用をする必要がないようにするため)
    eval "$(anyenv init -)"
  fi
fi
