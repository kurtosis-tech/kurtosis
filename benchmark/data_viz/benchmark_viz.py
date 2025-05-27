
import os
import re
import sys
import math
import pandas as pd
import matplotlib.pyplot as plt

if len(sys.argv) != 2:
    print("Usage: python benchmark_viz_single_image.py <data_directory>")
    sys.exit(1)

DATA_DIR = sys.argv[1]

def parse_duration(duration_str):
    unit_multipliers = {'s': 1, 'ms': 1e-3, 'µs': 1e-6, 'ns': 1e-9}
    match = re.match(r"([\d.]+)([a-zµ]+)", str(duration_str))
    if not match:
        return 0
    value, unit = match.groups()
    return float(value) * unit_multipliers.get(unit, 1)

# Collect plots
plots = []

def add_plot(title, plot_func):
    plots.append((title, plot_func))

def plot_add_services(df, ax):
    df["Add Time (s)"] = df["Time To Add Service Container"].apply(parse_duration)
    df["Readiness Time (s)"] = df["Time To Readiness Check"].apply(parse_duration)
    df.plot(x="Service Name", y=["Add Time (s)", "Readiness Time (s)"], kind="bar", ax=ax, legend=True)
    ax.set_title("Add Services Benchmark")
    ax.set_xlabel("Service Name")
    ax.tick_params(axis='x', labelrotation=45)

def plot_startosis(df, ax):
    df["Duration (s)"] = df["Value"].apply(parse_duration)
    df.plot(x="Metric", y="Duration (s)", kind="bar", ax=ax, legend=False)
    ax.set_title("Startosis Benchmark")
    ax.set_xlabel("Metric")
    ax.set_ylabel("Duration (s)")
    ax.tick_params(axis='x', labelrotation=45)

def plot_kurtosis_plan(df, ax):
    df["Total Time (s)"] = df["Total Time in Instruction"].apply(parse_duration)
    df["Number of Instructions"] = df["Number of Instructions"].astype(int)

    ax2 = ax.twinx()
    ax.bar(df["Instruction Name"], df["Total Time (s)"], color='tab:blue', alpha=0.7)
    ax.set_ylabel("Total Time (s)", color='tab:blue')
    ax.tick_params(axis='y', labelcolor='tab:blue')

    ax2.plot(df["Instruction Name"], df["Number of Instructions"], color='tab:orange', marker='o')
    ax2.set_ylabel("Num Instructions", color='tab:orange')
    ax2.tick_params(axis='y', labelcolor='tab:orange')

    ax.set_title("Kurtosis Plan Instructions")
    ax.set_xticks(range(len(df["Instruction Name"])))
    ax.set_xticklabels(df["Instruction Name"], rotation=45, ha="right")

def plot_run_sh(df, ax):
    df["Add Task Time (s)"] = df["Time To Add Task Container"].apply(parse_duration)
    df["Exec Time (s)"] = df["Time To Exec With Wait"].apply(parse_duration)
    df.plot(x="Task Name", y=["Add Task Time (s)", "Exec Time (s)"], kind="bar", ax=ax)
    ax.set_title("Run.sh Benchmark")
    ax.set_xlabel("Task Name")
    ax.tick_params(axis='x', labelrotation=45)

# Dispatch based on headers
for filename in os.listdir(DATA_DIR):
    if not filename.endswith(".csv"):
        continue
    filepath = os.path.join(DATA_DIR, filename)
    df = pd.read_csv(filepath)
    headers = set(df.columns)
    base = filename.replace(".csv", "")

    try:
        if {"Service Name", "Time To Add Service Container", "Time To Readiness Check"}.issubset(headers):
            add_plot(base, lambda ax, df=df: plot_add_services(df, ax))
        elif {"Metric", "Value"}.issubset(headers):
            add_plot(base, lambda ax, df=df: plot_startosis(df, ax))
        elif {"Instruction Name", "Total Time in Instruction", "Number of Instructions"}.issubset(headers):
            add_plot(base, lambda ax, df=df: plot_kurtosis_plan(df, ax))
        elif {"Task Name", "Time To Add Task Container", "Time To Exec With Wait"}.issubset(headers):
            add_plot(base, lambda ax, df=df: plot_run_sh(df, ax))
        else:
            print(f"Unrecognized CSV format: {filename}")
    except Exception as e:
        print(f"Error processing {filename}: {e}")

# Determine layout
num_plots = len(plots)
cols = 2
rows = math.ceil(num_plots / cols)
fig, axes = plt.subplots(rows, cols, figsize=(cols * 8, rows * 6))

# Normalize axes to 2D list
if num_plots == 1:
    axes = [[axes]]
elif rows == 1:
    axes = [axes]
elif cols == 1:
    axes = [[ax] for ax in axes]


# Draw plots and save individual subplot images
output_dir = "{0}/visualizations".format(DATA_DIR)
os.makedirs(output_dir, exist_ok=True)

for idx, (title, plot_func) in enumerate(plots):
    r, c = divmod(idx, cols)
    ax = axes[r][c]
    
plot_func(ax)
fig_individual, ax_ind = plt.subplots(figsize=(8, 6))
plot_func(ax_ind)
fig_individual.tight_layout()
fig_individual.savefig(os.path.join(output_dir, f"{title}_plot.png"))
plt.close(fig_individual)


# Hide any unused axes
for idx in range(num_plots, rows * cols):
    r, c = divmod(idx, cols)
    fig.delaxes(axes[r][c])

fig.tight_layout()
fig.suptitle("Benchmark Visualizations", fontsize=16, y=1.02)
plt.savefig("benchmark_visualizations_combined.png", bbox_inches="tight")
