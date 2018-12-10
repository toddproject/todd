git fetch origin master
if echo $(git diff --name-only master) | grep -w CHANGELOG.md > /dev/null; then
    echo "Thanks for making a CHANGELOG update!"
    exit 0
else
    echo "No CHANGELOG update found. Please provide update to CHANGELOG for this change."
    exit 1
fi