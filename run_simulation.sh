#!/bin/bash

# Define the lists
datasets=("femnist" "shakespeare")
algorithms=("fedavg" "minibatch")
nclients=(5 10 30 50)
minibatch_vals=(0.1 0.2 0.5 0.8)

# Preprocess vars
output_dir="trace_driven_simulator/data/"
device_flops=8

# Prepocessing data
if [ ! -d "$output_dir" ]; then
  for dataset in "${datasets[@]}"; do
    python3 trace_driven_simulator/data_processor.py --sample-dir "leaf_output/${dataset}/sys/" --search-pattern "metrics_sys_*" --output-dir "${output_dir}"
  done
fi

# Run simulations
for dataset in "${datasets[@]}"; do
  for algorithm in "${algorithms[@]}"; do
    if [ "$algorithm" == "minibatch" ]; then
        for minibatch_val in "${minibatch_vals[@]}"; do
            trace_file=`echo "trace_driven_simulator/data/metrics_sys_${dataset}_${algorithm}_c_30_mb_${minibatch_val}.csv"`
            go run trace_driven_simulator/main.go -t "${trace_file}"
        done
    else
        for nclient in "${nclients[@]}"; do
            trace_file=`echo "trace_driven_simulator/data/metrics_sys_${dataset}_${algorithm}_c_${nclient}_e_1.csv"`
            go run trace_driven_simulator/main.go -t "${trace_file}"
        done
    fi
  done
done
