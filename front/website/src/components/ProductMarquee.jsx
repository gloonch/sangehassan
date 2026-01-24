export default function ProductMarquee({ items }) {
  const trackItems = [...items, ...items];
  return (
    <div className="mask-fade overflow-hidden">
      <div className="flex w-[200%] animate-marquee gap-6">
        {trackItems.map((item, index) => (
          <div
            key={`${item}-${index}`}
            className="glass-panel flex min-w-[220px] flex-col gap-3 rounded-2xl p-5"
          >
            <div className="h-28 rounded-xl bg-primary/10" />
            <p className="text-sm font-semibold text-primary">{item}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
