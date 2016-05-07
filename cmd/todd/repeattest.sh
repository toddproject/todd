while true
do
sleep 5
todd run test-ping-dns-dc -j -y
sleep 5
todd run test-ping-dns-hq -j -y
sleep 5
todd run test-dc-hq-bandwidth -j -y
done

