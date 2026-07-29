# Oh My Zsh Plugin for SKM (Skill Manager)

# Add plugin directory to fpath for zsh completion
if [[ ! ${fpath[(r)${0:h}]} ]]; then
    fpath=(${0:h} $fpath)
fi

# Handy skm Command Aliases
alias skma="skm add"
alias skmv="skm validate"
alias skml="skm list"
alias skms="skm search"
alias skmc="skm compile"
alias skmi="skm init"
