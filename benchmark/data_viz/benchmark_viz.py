
import os
import re
import pandas as pd
import matplotlib.pyplot as plt

import sys
if len(sys.argv) != 2:
    print("Usage: python benchmark_viz_customized.py <data_directory>")
    sys.exit(1)
DATA_DIR = sys.argv[1]  # Change this to your benchmark CSV directory
os.makedirs("visualizations", exist_ok=True)

def parse_duration(duration_str):
    unit_multipliers = {'s': 1, 'ms': 1e-3, 'µs': 1e-6, 'ns': 1e-9}
    match = re.match(r"([\d.]+)([a-zµ]+)", str(duration_str))
    if not match:
        return 0
    value, unit = match.groups()
    return float(value) * unit_multipliers.get(unit, 1)

def plot_add_services(df, filename):
    df["Add Time (s)"] = df["Time To Add Service Container"].apply(parse_duration)
    df["Readiness Time (s)"] = df["Time To Readiness Check"].apply(parse_duration)
    df.plot(x="Service Name", y=["Add Time (s)", "Readiness Time (s)"], kind="bar", figsize=(12, 6))
    plt.title("Add Services Benchmark")
    plt.xticks(rotation=45, ha="right")
    plt.tight_layout()
    plt.savefig(f"visualizations/{filename}_plot.png")
    plt.close()

def plot_startosis(df, filename):
    df["Duration (s)"] = df["Value"].apply(parse_duration)
    df.plot(x="Metric", y="Duration (s)", kind="bar", legend=False, figsize=(10, 6))
    plt.title("Startosis Benchmark")
    plt.xticks(rotation=45, ha="right")
    plt.ylabel("Duration (s)")
    plt.tight_layout()
    plt.savefig(f"visualizations/{filename}_plot.png")
    plt.close()

def plot_kurtosis_plan(df, filename):
    df["Total Time (s)"] = df["Total Time in Instruction"].apply(parse_duration)
    df["Number of Instructions"] = df["Number of Instructions"].astype(int)
    fig, ax1 = plt.subplots(figsize=(12, 6))

    color1 = "tab:blue"
    color2 = "tab:orange"
    ax1.set_xlabel("Instruction Name")
    ax1.set_ylabel("Total Time (s)", color=color1)
    ax1.bar(df["Instruction Name"], df["Total Time (s)"], color=color1, alpha=0.7)
    ax1.tick_params(axis="y", labelcolor=color1)
    ax1.set_xticks(range(len(df["Instruction Name"])))
    ax1.set_xticklabels(df["Instruction Name"], rotation=45, ha="right")

    ax2 = ax1.twinx()
    ax2.set_ylabel("Number of Instructions", color=color2)
    ax2.plot(df["Instruction Name"], df["Number of Instructions"], color=color2, marker='o')
    ax2.tick_params(axis="y", labelcolor=color2)

    fig.tight_layout()
    plt.title("Kurtosis Plan Instructions Benchmark")
    plt.savefig(f"visualizations/{filename}_plot.png")
    plt.close()

def plot_run_sh(df, filename):
    df["Add Task Time (s)"] = df["Time To Add Task Container"].apply(parse_duration)
    df["Exec Time (s)"] = df["Time To Exec With Wait"].apply(parse_duration)
    df.plot(x="Task Name", y=["Add Task Time (s)", "Exec Time (s)"], kind="bar", figsize=(12, 6))
    plt.title("Run.sh Benchmark")
    plt.xticks(rotation=45, ha="right")
    plt.tight_layout()
    plt.savefig(f"visualizations/{filename}_plot.png")
    plt.close()

# Dispatch based on headers
for filename in os.listdir(DATA_DIR):
    if not filename.endswith(".csv"):
        continue
    filepath = os.path.join(DATA_DIR, filename)
    df = pd.read_csv(filepath)
    headers = set(df.columns)

    try:
        if {"Service Name", "Time To Add Service Container", "Time To Readiness Check"}.issubset(headers):
            plot_add_services(df, filename.replace(".csv", ""))
        elif {"Metric", "Value"}.issubset(headers):
            plot_startosis(df, filename.replace(".csv", ""))
        elif {"Instruction Name", "Total Time in Instruction", "Number of Instructions"}.issubset(headers):
            plot_kurtosis_plan(df, filename.replace(".csv", ""))
        elif {"Task Name", "Time To Add Task Container", "Time To Exec With Wait"}.issubset(headers):
            plot_run_sh(df, filename.replace(".csv", ""))
        else:
            print(f"Unrecognized CSV format: {filename}")
    except Exception as e:
        print(f"Error processing {filename}: {e}")
