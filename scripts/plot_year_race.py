#!/usr/bin/env python3
"""
Plot year/race_int vs race_int/price from the cars table using Plotly.
Requires .env with DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME.
Outputs: plot_year_race.html, plot_year_price.html, plot_race_price.html
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

PCTL_HI = 0.90
PCTL_LO = 0.10
SMOOTH_WINDOW = 5


def percentile(sorted_arr, p):
    """p in 0..1; returns value at that percentile."""
    if not sorted_arr:
        return None
    idx = min(int(p * (len(sorted_arr) - 1)), len(sorted_arr) - 1)
    return sorted_arr[idx]


def moving_avg(values, window=SMOOTH_WINDOW):
    """Smooth with centered moving average; window must be odd."""
    n = len(values)
    half = window // 2
    out = []
    for i in range(n):
        lo = max(0, i - half)
        hi = min(n, i + half + 1)
        out.append(sum(values[lo:hi]) / (hi - lo))
    return out


def smooth_max_min(x_vals, y_vals, group_key_fn=None, x_label_fn=None):
    """
    Group (x_vals, y_vals) by key = group_key_fn(x) (default: x itself).
    For each group compute 90th and 10th percentile of y; then moving average.
    Returns (sorted_x_labels, smooth_max_list, smooth_min_list).
    x_label_fn: optional, key -> label for x axis (e.g. bin center); default identity.
    """
    if group_key_fn is None:
        group_key_fn = lambda x: x
    if x_label_fn is None:
        x_label_fn = lambda k: k
    grouped = {}
    for x, y in zip(x_vals, y_vals):
        k = group_key_fn(x)
        grouped.setdefault(k, []).append(y)
    sorted_keys = sorted(grouped.keys())
    robust_max = [percentile(sorted(grouped[k]), PCTL_HI) for k in sorted_keys]
    robust_min = [percentile(sorted(grouped[k]), PCTL_LO) for k in sorted_keys]
    x_labels = [x_label_fn(k) for k in sorted_keys]
    return x_labels, moving_avg(robust_max), moving_avg(robust_min)


def add_scatter_with_smooth_lines(fig, x_vals, y_vals, x_title, y_title, title, group_key_fn=None, x_label_fn=None):
    """Add scatter + smooth max/min lines to fig (single plot)."""
    xs, smooth_max, smooth_min = smooth_max_min(x_vals, y_vals, group_key_fn, x_label_fn)
    fig.add_trace(go.Scatter(x=x_vals, y=y_vals, mode="markers", name="cars", opacity=0.5, marker=dict(size=4)))
    fig.add_trace(go.Scatter(x=xs, y=smooth_max, mode="lines", name="smooth max (90th pctl)", line=dict(color="red", width=2)))
    fig.add_trace(go.Scatter(x=xs, y=smooth_min, mode="lines", name="smooth min (10th pctl)", line=dict(color="blue", width=2)))
    fig.update_layout(title=title, xaxis_title=x_title, yaxis_title=y_title, template="plotly_white", showlegend=True)
    fig.update_layout(margin=dict(l=60, r=40, t=50, b=50))


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

    # 1) Year vs race_int
    cur.execute("SELECT year, race_int FROM cars WHERE year > 0 AND race_int >= 0 ORDER BY year, race_int")
    rows_yr = cur.fetchall()
    if not rows_yr:
        print("No rows for year/race_int", file=sys.stderr)
        cur.close()
        conn.close()
        sys.exit(1)
    years = [r[0] for r in rows_yr]
    race_ints = [r[1] for r in rows_yr]
    fig1 = go.Figure()
    add_scatter_with_smooth_lines(fig1, years, race_ints, "Year", "race_int", "Year vs Mileage (race_int)")
    fig1.update_layout(xaxis=dict(dtick=1))
    out1 = os.path.join(os.path.dirname(__file__), "plot_year_race.html")
    fig1.write_html(out1)
    print(f"Wrote {out1}")

    # 2) Year vs price (usd)
    cur.execute("SELECT year, usd FROM cars WHERE year > 0 AND usd > 0 ORDER BY year, usd")
    rows_yp = cur.fetchall()
    if rows_yp:
        years_p = [r[0] for r in rows_yp]
        prices = [r[1] for r in rows_yp]
        fig2 = go.Figure()
        add_scatter_with_smooth_lines(fig2, years_p, prices, "Year", "price (USD)", "Year vs Price (USD)")
        fig2.update_layout(xaxis=dict(dtick=1))
        out2 = os.path.join(os.path.dirname(__file__), "plot_year_price.html")
        fig2.write_html(out2)
        print(f"Wrote {out2}")
    else:
        print("No rows for year/price", file=sys.stderr)

    # 3) race_int vs price (usd) — bin mileage so we have enough points per group
    cur.execute("SELECT race_int, usd FROM cars WHERE race_int > 0 AND usd > 0 ORDER BY race_int, usd")
    rows_rp = cur.fetchall()
    if rows_rp:
        BIN_STEP = 20000  # 20k km bins
        race_vals = [r[0] for r in rows_rp]
        price_vals = [r[1] for r in rows_rp]
        group_key = lambda x: (x // BIN_STEP) * BIN_STEP
        bin_center = lambda k: k + BIN_STEP // 2
        fig3 = go.Figure()
        add_scatter_with_smooth_lines(
            fig3, race_vals, price_vals,
            "race_int (mileage)", "price (USD)", "Mileage (race_int) vs Price (USD)",
            group_key_fn=group_key, x_label_fn=bin_center,
        )
        out3 = os.path.join(os.path.dirname(__file__), "plot_race_price.html")
        fig3.write_html(out3)
        print(f"Wrote {out3}")
    else:
        print("No rows for race_int/price", file=sys.stderr)

    cur.close()
    conn.close()


if __name__ == "__main__":
    main()
