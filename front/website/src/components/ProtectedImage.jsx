const defaultImgClass =
  "h-full w-full select-none [-webkit-touch-callout:none] [-webkit-user-drag:none]";

const stopImageEvent = (event) => {
  event.preventDefault();
};

const stopClickEvent = (event) => {
  event.preventDefault();
  event.stopPropagation();
};

export default function ProtectedImage({
  src,
  alt = "",
  className = "",
  wrapperClassName = "",
  imageClassName = "",
  width,
  height,
  loading,
  decoding = "async",
  fetchPriority,
  fit = "cover",
  stopClicks = false,
  ...props
}) {
  const fitClassName = fit === "contain" ? "object-contain" : "object-cover";
  const imgClassName = [defaultImgClass, fitClassName, className, imageClassName].filter(Boolean).join(" ");

  return (
    <span
      data-protected-image="true"
      className={`relative block select-none overflow-hidden ${wrapperClassName}`}
      onContextMenu={stopImageEvent}
      onDragStart={stopImageEvent}
      onClickCapture={stopClicks ? stopClickEvent : undefined}
    >
      <img
        {...props}
        src={src}
        alt={alt}
        width={width}
        height={height}
        loading={loading}
        decoding={decoding}
        fetchpriority={fetchPriority}
        draggable={false}
        className={imgClassName}
      />
      <span aria-hidden="true" className="pointer-events-none absolute inset-0 select-none" />
    </span>
  );
}
