#!/bin/zsh

# direnvの有効化
if ! grep -q "direnv hook zsh" ~/.zshrc; then
  if command -v direnv 2>&1 >/dev/null; then
    # .zshrchへの書き込み
    echo "eval \"\$(direnv hook zsh)\"" >> ~/.zshrc
    # direnvの有効化(.zshrcの書き込みした直後は.zshrcの再適用をする必要がないようにするため)
    eval "$(direnv hook zsh)"
  fi
fi
