mkdir -p tree
rm -rf tree/*



for i in {1..5}
do
    mkdir -p tree/folder_$i
    for j in {1..10}
    do
        mkdir -p tree/folder_$i/folder_$j
        for k in {1..25}
        do
            mkdir -p tree/folder_$i/folder_$j/file_$k
        done
    done
done
