const VIEWBOX_WIDTH = 200;
const VIEWBOX_HEIGHT = 36;
const STROKE_PADDING = 1.5;

type SparklineProps = {
  values: number[];
  stroke: string;
  variant?: "line" | "bars";
};

/**
 * Axis-less inline trend for metric cards. Plain SVG in its own module so the
 * KPI row does not pull recharts into the dashboard's first-load bundle;
 * `dashboard-charts.tsx` re-exports it.
 */
export function Sparkline({ values, stroke, variant = "line" }: SparklineProps) {
  const finite = values.filter((value) => Number.isFinite(value));
  if (finite.length === 0) return null;

  const max = Math.max(...finite);
  const min = variant === "bars" ? 0 : Math.min(...finite);
  const span = max - min || 1;
  const usableHeight = VIEWBOX_HEIGHT - STROKE_PADDING * 2;
  const toY = (value: number) => STROKE_PADDING + usableHeight - ((value - min) / span) * usableHeight;

  return (
    <svg
      viewBox={`0 0 ${VIEWBOX_WIDTH} ${VIEWBOX_HEIGHT}`}
      preserveAspectRatio="none"
      className="block h-9 w-full"
      aria-hidden="true"
      data-slot="sparkline"
    >
      {variant === "bars" ? (
        <SparkBars values={finite} fill={stroke} toY={toY} />
      ) : (
        <polyline
          points={finite
            .map((value, index) => {
              const x = finite.length === 1 ? VIEWBOX_WIDTH / 2 : (index / (finite.length - 1)) * VIEWBOX_WIDTH;
              return `${x.toFixed(2)},${toY(value).toFixed(2)}`;
            })
            .join(" ")}
          fill="none"
          stroke={stroke}
          strokeWidth={1.5}
          strokeLinejoin="round"
          strokeLinecap="round"
          vectorEffect="non-scaling-stroke"
        />
      )}
    </svg>
  );
}

function SparkBars({ values, fill, toY }: { values: number[]; fill: string; toY: (value: number) => number }) {
  const slot = VIEWBOX_WIDTH / values.length;
  const barWidth = slot * 0.6;
  return (
    <>
      {values.map((value, index) => {
        const y = toY(value);
        return (
          <rect
            key={index}
            x={index * slot + (slot - barWidth) / 2}
            y={y}
            width={barWidth}
            height={Math.max(0, VIEWBOX_HEIGHT - y)}
            rx={1.5}
            fill={fill}
            opacity={index === values.length - 1 ? 1 : 0.55}
          />
        );
      })}
    </>
  );
}
