#!/bin/bash
# This script will be sourced with the environment file path as the first argument
source "$1"
SESSION_DIR=$BASE_DIR/$SESSION_ID
REPO_URL=https://x-oauth-token:${GITHUB_TOKEN}@github.com/$REPO
mkdir -p $SESSION_DIR
if [ -d "$SESSION_DIR/repo" ]; then
        cd $SESSION_DIR/repo
        git remote remove origin
        git remote add origin $REPO_URL
        git fetch origin
else
        mkdir -p $SESSION_DIR
        git clone $REPO_URL $SESSION_DIR/repo
        cd $SESSION_DIR/repo
fi

# Git configuration - these values will be passed as arguments 2 and 3
git config --global user.email "$2"
git config --global user.name "$3"

# PRブランチが指定されている場合はそのブランチをチェックアウト
if [ -n "$PR_BRANCH" ]; then
        echo "Checking out PR branch: $PR_BRANCH"
        git checkout $PR_BRANCH || git checkout -b $PR_BRANCH origin/$PR_BRANCH
fi

run_goose() {
        RESUME=$1
        goose run --name $SESSION_ID $RESUME \
                --with-builtin "developer" \
                --with-extension "GITHUB_PERSONAL_ACCESS_TOKEN=$GITHUB_TOKEN mise exec -- npx -y @modelcontextprotocol/server-github" \
                --with-extension "MEMORY_BANK_ROOT=$HOME/.kommon/memory mise exec -- npx -y @allpepper/memory-bank-mcp" \
                --with-extension "mise exec -- npx -y @modelcontextprotocol/server-sequential-thinking" \
                --instructions $INSTRUCTION_FILE_PATH
        return $?
}
run_goose -r || run_goose
wait