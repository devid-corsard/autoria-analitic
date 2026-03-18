#!/usr/bin/env python3
"""
Plot year (x) vs race_int (y) from the cars table using Plotly.
Requires .env with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME.
Output: plot_year_race.html
"""
import os
import sys

import plotly.graph_objects as go
import psycopg2
from dotenv import load_dotenv

# Load .env from project root (parent of scripts/)
load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), "..", ".env"))

DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = os.getenv("DB_PORT", "5432")
DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")
DB_NAME = os.getenv("DB_NAME")


def main():
    if not all([DB_USER, DB_PASSWORD, DB_NAME]):
        print("Set DB_USER, DB_PASSWORD, DB_NAME in .env", file=sys.stderr)
        sys.exit(1)

    conn = psycopg2.connect(
        host=DB_HOST,
        port=DB_PORT,
        user=DB_USER,
        password=DB_PASSWORD,
        dbname=DB_NAME,
    )
    cur = conn.cursor()
    cur.execute("SELECT year, race_int FROM cars WHERE year > 0 AND race_int >= 0 ORDER BY year, race_int")
    rows = cur.fetchall()
    cur.close()
    conn.close()

    if not rows:
        print("No rows with year > 0 and race_int >= 0", file=sys.stderr)
        sys.exit(1)

    years = [r[0] for r in rows]
    race_ints = [r[1] for r in rows]

    # Group by year and compute robust max/min per year (percentiles to ignore outliers)
    by_year = {}
    for y, r in zip(years, race_ints):
        by_year.setdefault(y, []).append(r)

    def percentile(sorted_arr, p):
        """p in 0..1; returns value at that percentile."""
        if not sorted_arr:
            return None
        idx = min(int(p * (len(sorted_arr) - 1)), len(sorted_arr) - 1)
        return sorted_arr[idx]

    # Robust max = 90th percentile, robust min = 10th percentile (ignore extremes)
    sorted_years = sorted(by_year.keys())
    robust_max = []
    robust_min = []
    for y in sorted_years:
        vals = sorted(by_year[y])
        robust_max.append(percentile(vals, 0.90))
        robust_min.append(percentile(vals, 0.10))

    def moving_avg(values, window=5):
        """Smooth with centered moving average; window must be odd."""
        n = len(values)
        half = window // 2
        out = []
        for i in range(n):
            lo = max(0, i - half)
            hi = min(n, i + half + 1)
            out.append(sum(values[lo:hi]) / (hi - lo))
        return out

    smooth_max = moving_avg(robust_max)
    smooth_min = moving_avg(robust_min)

    fig = go.Figure(
        data=[
            go.Scatter(x=years, y=race_ints, mode="markers", name="cars", opacity=0.5, marker=dict(size=4)),
            go.Scatter(x=sorted_years, y=smooth_max, mode="lines", name="smooth max (90th pctl)", line=dict(color="red", width=2)),
            go.Scatter(x=sorted_years, y=smooth_min, mode="lines", name="smooth min (10th pctl)", line=dict(color="blue", width=2)),
        ],
        layout=go.Layout(
            title="Year vs Mileage (race_int)",
            xaxis_title="Year",
            yaxis_title="race_int",
            template="plotly_white",
            showlegend=True,
        ),
    )
    fig.update_layout(
        xaxis=dict(dtick=1),
        margin=dict(l=60, r=40, t=50, b=50),
    )

    out = os.path.join(os.path.dirname(__file__), "plot_year_race.html")
    fig.write_html(out)
    print(f"Wrote {out}")


if __name__ == "__main__":
    main()
