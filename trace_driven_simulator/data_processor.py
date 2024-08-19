from os import listdir, path, makedirs
import re
import pandas as pd

def find_leaf_stats_files(directory:str, pattern:str) -> list[str]:
    regex = re.compile(pattern)

    stats_files:set[str] = set()

    for filename in listdir(directory):
        if filename.endswith('.csv') and regex.match(filename):
            stats_files.add(filename)
    
    return list(stats_files)


def main():
    SAMPLE_DIR_NAME="leaf_output/femnist/sys/"
    OUTPUT_DIR_NAME="trace_driven_simulator/data/"
    LEAF_STATS_REGEX=r'metrics_sys_*'
    
    # Cenário Otimista
    CROSSILO_PROCESSOR_THROUGHPUT = 15000 * 10**9 # Assumindo uma NVIDIA Tesla
    CROSSDEVICE_PROCESSOR_THROUGHPUT = 700 * 10**9 # Assumindo um Adreno 660

    print("Colleting system metrics from the given directory...")

    try:
        sys_stats_filenames = find_leaf_stats_files(SAMPLE_DIR_NAME, LEAF_STATS_REGEX)
    except Exception as error:
        print(f"Unable to find expected files: {error}")

    print("Starting data processing")

    for sys_stats in sys_stats_filenames:
        df = pd.read_csv(
            SAMPLE_DIR_NAME + sys_stats,
            names=[
                "client_id",
                "round_number",
                "hierarchy",
                "num_samples",
                "set",
                "bytes_written",
                "bytes_sended",
                "local_computations"
            ]
        )

        client_id_map = {}  # Dictionary to store the mapping of client_id to integers
        client_id_counter = 1  # Counter to assign unique integer values to client_ids

        # A primeira coluna é sempre vazia e a segunda é igual a bytes_sended
        df = df.drop(["hierarchy", "bytes_written"], axis=1)

        # Map client_id to unique integers based on round_number
        df['client_id'] = df.groupby(['round_number'])['client_id'].transform(
            lambda x: pd.factorize(x)[0] + 1
        )

        splits = sys_stats.split("_")

        # response time in miliseconds
        if int(splits[5]) >= 10:
            df['time'] = (df['local_computations'] / CROSSDEVICE_PROCESSOR_THROUGHPUT)
        else:
            df['time'] = (df['local_computations'] / CROSSILO_PROCESSOR_THROUGHPUT)
        
        # Prepare output file path
        output_file_path = path.join(OUTPUT_DIR_NAME, path.basename(sys_stats))

        # Ensure the output directory exists
        makedirs(OUTPUT_DIR_NAME, exist_ok=True)

        # Save the modified dataframe to CSV in the output directory
        df.to_csv(output_file_path, index=False)

        print(f"Data from {sys_stats} was processed")

if __name__ == "__main__":
    main()