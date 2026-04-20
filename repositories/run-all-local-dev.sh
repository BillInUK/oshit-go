#!/bin/bash
# 按顺序定义文件列表
FILES=(
    "structure/base_structure.sql"
    "structure/reward_structure.sql"
    "structure/pos_structure.sql"
    "structure/stake_structure.sql"
    "config/testnet_base_config.sql"
    "config/testnet_reward_config.sql"
    "config/testnet_pos_config.sql"
    "config/testnet_stake_config.sql"
    "test_data/testnet_data.sql"
)

for file in "${FILES[@]}"; do
    echo "Executing $file ..."
    docker exec -i postgres16.2 psql -U postgres -d oshit_db < "$file"
    if [ $? -ne 0 ]; then
        echo "Error executing $file, stopping."
        exit 1
    fi
done
echo "All SQL files executed successfully."