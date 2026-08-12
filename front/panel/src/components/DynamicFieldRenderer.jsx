import { useRef, useState } from "react";

const measurementTypes = new Set(["WEIGHT", "AREA", "VOLUME", "QUANTITY"]);

function SignaturePad({ onUpload, disabled }) {
  const canvasRef = useRef(null);
  const [drawing, setDrawing] = useState(false);
  const point = (event) => {
    const canvas = canvasRef.current;
    const rect = canvas.getBoundingClientRect();
    return { x: (event.clientX - rect.left) * (canvas.width / rect.width), y: (event.clientY - rect.top) * (canvas.height / rect.height) };
  };
  const start = (event) => {
    if (disabled) return;
    const context = canvasRef.current.getContext("2d");
    const p = point(event);
    context.beginPath(); context.moveTo(p.x, p.y); setDrawing(true);
  };
  const draw = (event) => {
    if (!drawing) return;
    const context = canvasRef.current.getContext("2d");
    const p = point(event); context.lineWidth = 2; context.lineCap = "round"; context.lineTo(p.x, p.y); context.stroke();
  };
  const save = () => canvasRef.current.toBlob((blob) => onUpload(new File([blob], "signature.png", { type: "image/png" })), "image/png");
  return <div className="space-y-2"><canvas ref={canvasRef} width="640" height="180" onPointerDown={start} onPointerMove={draw} onPointerUp={()=>setDrawing(false)} onPointerLeave={()=>setDrawing(false)} className="h-36 w-full touch-none rounded-xl border bg-white"/><div className="flex gap-2"><button type="button" disabled={disabled} onClick={save} className="rounded-full border px-3 py-1 text-xs">ثبت امضا</button><button type="button" disabled={disabled} onClick={()=>canvasRef.current.getContext("2d").clearRect(0,0,640,180)} className="rounded-full border px-3 py-1 text-xs">پاک کردن</button></div><p className="text-xs text-primary/50">این تصویر صرفاً ثبت داخلی است و امضای حقوقی محسوب نمی‌شود.</p></div>;
}

export default function DynamicFieldRenderer({ field, value, onChange, onUpload, disabled = false }) {
  const common = { disabled, required: field.is_required, className: "w-full rounded-xl border bg-white p-3", value: value ?? "", onChange: (event) => onChange(event.target.value) };
  const options = Array.isArray(field.options_json) ? field.options_json : [];
  const optionValue = (option) => typeof option === "string" ? option : option.value;
  const optionLabel = (option) => typeof option === "string" ? option : (option.label_fa || option.label || option.value);
  let control;
  if (field.field_type === "BOOLEAN") control = <input type="checkbox" disabled={disabled} checked={Boolean(value)} onChange={event=>onChange(event.target.checked)}/>;
  else if (field.field_type === "LONG_TEXT" || field.field_type === "ADDRESS") control = <textarea {...common} rows="4"/>;
  else if (["INTEGER", "DECIMAL"].includes(field.field_type)) control = <input {...common} type="number" step={field.field_type === "INTEGER" ? "1" : "any"} onChange={event=>onChange(event.target.value === "" ? "" : Number(event.target.value))}/>;
  else if (["DATE", "TIME", "DATETIME"].includes(field.field_type)) control = <input {...common} type={field.field_type === "DATETIME" ? "datetime-local" : field.field_type.toLowerCase()}/>;
  else if (field.field_type === "SELECT") control = <select {...common}><option value="">انتخاب کنید</option>{options.map(option=><option key={optionValue(option)} value={optionValue(option)}>{optionLabel(option)}</option>)}</select>;
  else if (field.field_type === "MULTI_SELECT") control = <div className="flex flex-wrap gap-3">{options.map(option=>{const optionKey=optionValue(option);return <label key={optionKey} className="flex items-center gap-2 text-sm"><input type="checkbox" disabled={disabled} checked={(value||[]).includes(optionKey)} onChange={event=>onChange(event.target.checked?[...(value||[]),optionKey]:(value||[]).filter(item=>item!==optionKey))}/>{optionLabel(option)}</label>})}</div>;
  else if (measurementTypes.has(field.field_type)) control = <div className="flex"><input {...common} className="w-full rounded-r-xl border p-3" type="number" step="any" value={value?.value ?? ""} onChange={event=>onChange({value:event.target.value===""?"":Number(event.target.value),unit:field.unit_code})}/><span className="rounded-l-xl border border-r-0 bg-primary/5 px-4 py-3">{field.unit_code}</span></div>;
  else if (field.field_type === "MONEY") control = <div className="flex"><input {...common} className="w-full rounded-r-xl border p-3" type="number" step="any" value={value?.amount ?? ""} onChange={event=>onChange({amount:event.target.value===""?"":Number(event.target.value),currency:field.currency_code||"IRR"})}/><span className="rounded-l-xl border border-r-0 bg-primary/5 px-4 py-3">{field.currency_code||"IRR"}</span></div>;
  else if (field.field_type === "QC_CHECK") {
    const qc = value && typeof value === "object" ? value : {};
    const updateQC = (changes) => onChange({ ...qc, ...changes });
    control = <div className="space-y-3 rounded-xl border bg-primary/5 p-3">
      <select disabled={disabled} required={field.is_required} className="w-full rounded-xl border bg-white p-3" value={qc.result || ""} onChange={event=>updateQC({result:event.target.value})}>
        <option value="">نتیجه را انتخاب کنید</option><option value="PASS">قبول</option><option value="FAIL">رد</option><option value="NOT_APPLICABLE">قابل اعمال نیست</option>
      </select>
      <div className="grid gap-2 sm:grid-cols-2"><input disabled={disabled} className="rounded-xl border bg-white p-3" inputMode="decimal" placeholder="مقدار اندازه‌گیری" value={qc.measuredValue ?? ""} onChange={event=>updateQC({measuredValue:event.target.value})}/><input disabled={disabled} className="rounded-xl border bg-white p-3" placeholder="واحد" value={qc.unit ?? ""} onChange={event=>updateQC({unit:event.target.value})}/></div>
      <textarea disabled={disabled} className="w-full rounded-xl border bg-white p-3" rows="2" placeholder="یادداشت کنترل کیفیت" value={qc.note ?? ""} onChange={event=>updateQC({note:event.target.value})}/>
      <input disabled={disabled} type="file" accept="image/png,image/jpeg" capture="environment" onChange={event=>event.target.files?.[0]&&onUpload(event.target.files[0])}/>
      {qc.fileId&&<a className="text-sm underline" href={`/api/v1/workflow-files/${qc.fileId}`} target="_blank" rel="noreferrer">مشاهده تصویر</a>}
    </div>;
  }
  else if (field.field_type === "SIGNATURE") control = <SignaturePad disabled={disabled} onUpload={onUpload}/>;
  else if (["FILE", "IMAGE"].includes(field.field_type)) control = <div><input disabled={disabled} type="file" accept={field.field_type==="IMAGE"?"image/png,image/jpeg":"image/png,image/jpeg,application/pdf"} onChange={event=>event.target.files?.[0]&&onUpload(event.target.files[0])}/>{value?.fileId&&<a className="mr-3 text-sm underline" href={`/api/v1/workflow-files/${value.fileId}`} target="_blank" rel="noreferrer">مشاهده فایل</a>}</div>;
  else control = <input {...common} type={field.field_type === "PHONE" ? "tel" : "text"} placeholder={field.placeholder_fa||""}/>;
  return <label className="block space-y-2"><span className="text-sm font-semibold">{field.label_fa}{field.is_required&&<b className="text-red-600"> *</b>}</span>{field.description_fa&&<small className="block text-primary/55">{field.description_fa}</small>}{control}</label>;
}
