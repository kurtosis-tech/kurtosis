import matplotlib.pyplot as plt
import pandas as pd
import re

# Load CSV
df = pd.read_csv("benchmark.csv")

# Convert duration to seconds
unit_multipliers = {
    's': 1,
    'ms': 1e-3,
    'µs': 1e-6,
    'ns': 1e-9,
}

def parse_duration(duration_str):
    match = re.match(r"([\d.]+)([a-zµ]+)", duration_str)
    if not match:
        return 0
    value, unit = match.groups()
    return float(value) * unit_multipliers.get(unit, 1)

df["Duration (s)"] = df["Duration"].apply(parse_duration)
df["Count"] = df["Count"].astype(int)

# Plot
fig, ax1 = plt.subplots(figsize=(12, 6))

color1 = 'tab:blue'
color2 = 'tab:orange'

ax1.set_title("Benchmark: Duration and Count per Instruction")
ax1.set_xlabel("Instruction")
ax1.set_ylabel("Duration (seconds)", color=color1)
bars1 = ax1.bar(df["Instruction"], df["Duration (s)"], color=color1, alpha=0.7, label="Duration")
ax1.tick_params(axis='y', labelcolor=color1)
ax1.set_xticks(range(len(df["Instruction"])))
ax1.set_xticklabels(df["Instruction"], rotation=45, ha='right')

# Scatter for counts
ax2 = ax1.twinx()
ax2.set_ylabel("Count", color=color2)
ax2.scatter(range(len(df["Instruction"])), df["Count"], color=color2, marker='o', label="Count")
ax2.tick_params(axis='y', labelcolor=color2)

fig.tight_layout()
plt.show()
