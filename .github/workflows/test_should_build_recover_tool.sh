if grep -q "build-recover-tool: true" cmd/recover/version; then
    echo 'result=true' >> $GITHUB_OUTPUT
else
    echo 'result=false' >> $GITHUB_OUTPUT
fi