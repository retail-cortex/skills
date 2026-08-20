# Oh My Zsh Plugin for Castor CLI (cstr)
 
# Add plugin directory to fpath for zsh completion
if [[ ! ${fpath[(r)${0:h}]} ]]; then
    fpath=(${0:h} $fpath)
fi

# Handy cstr Command Aliases
alias cstra="cstr add"
alias cstrver="cstr verify"
alias cstrv="cstr validate"
alias cstrl="cstr list"
alias cstrs="cstr search"
alias cstrc="cstr compile"
alias cstri="cstr init"

# Backwards compatibility aliases
alias skma="cstr add"
alias skmver="cstr verify"
alias skmv="cstr validate"
alias skml="cstr list"
alias skms="cstr search"
alias skmc="cstr compile"
alias skmi="cstr init"
alias skm="cstr"
