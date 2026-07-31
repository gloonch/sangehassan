import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { fetchJSON } from "../lib/api";

const statusFA={DRAFT:"پیش‌نویس",PUBLISHED:"منتشرشده",ARCHIVED:"آرشیوشده"};

export default function WorkflowTemplates(){
  const navigate=useNavigate();
  const [items,setItems]=useState([]),[open,setOpen]=useState(false),[error,setError]=useState("");
  const [form,setForm]=useState({template_group_code:"",name_fa:"",description_fa:"",icon_key:"workflow",start_permission_code:"",is_active:true});
  const load=async(isCancelled=()=>false)=>{
    try {
      const response=await fetchJSON("/api/v1/admin/workflow-templates");
      if(isCancelled())return;
      setItems(response.data||[]);
      setError("");
    } catch(loadError) {
      if(!isCancelled())setError(loadError.message);
    }
  };
  useEffect(()=>{
    let cancelled=false;
    void load(()=>cancelled);
    return()=>{cancelled=true};
  },[]);
  const create=async event=>{event.preventDefault();const response=await fetchJSON("/api/v1/admin/workflow-templates",{method:"POST",body:JSON.stringify(form)});navigate(`/dashboard/workflows/${response.data.id}/builder`)};
  const command=async(id,action)=>{await fetchJSON(`/api/v1/admin/workflow-templates/${id}/${action}`,{method:"POST"});void load()};
  const groups=items.reduce((all,item)=>({...all,[item.template_group_code]:[...(all[item.template_group_code]||[]),item]}),{});
  return <div className="space-y-5" dir="rtl"><section className="panel-card flex flex-wrap items-center justify-between gap-3"><div><h2 className="font-display text-2xl">نسخه‌های Workflow</h2><p className="text-sm text-primary/60">هر سفارش Snapshot مستقل نسخه منتشرشده را نگه می‌دارد.</p></div><button onClick={()=>setOpen(true)} className="rounded-full bg-primary px-5 py-2 text-sand">الگوی جدید</button></section>{error&&<p className="rounded-xl bg-red-50 p-4 text-red-700">{error}</p>}<div className="grid gap-4">{Object.entries(groups).map(([group,versions])=><section key={group} className="panel-card"><h3 dir="ltr" className="font-mono font-semibold">{group}</h3><div className="mt-4 overflow-x-auto"><table className="w-full text-sm"><thead><tr className="text-right text-primary/55"><th className="p-2">نسخه</th><th>عنوان</th><th>وضعیت</th><th>انتشار</th><th>عملیات</th></tr></thead><tbody>{versions.map(item=><tr key={item.id} className="border-t"><td className="p-2">v{item.version_number}</td><td>{item.name_fa}</td><td><span className="rounded-full bg-primary/5 px-2 py-1">{statusFA[item.status]}</span></td><td>{item.published_at?new Date(item.published_at).toLocaleDateString("fa-IR"):"—"}</td><td className="space-x-2 space-x-reverse"><Link className="underline" to={`/dashboard/workflows/${item.id}/builder`}>{item.status==="DRAFT"?"ویرایش":"مشاهده"}</Link>{item.status==="PUBLISHED"&&<><button className="underline" onClick={()=>command(item.id,"clone")}>ساخت نسخه جدید</button><button className="text-red-700 underline" onClick={()=>command(item.id,"archive")}>آرشیو</button></>}</td></tr>)}</tbody></table></div></section>)}</div>{open&&<div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"><form onSubmit={create} className="w-full max-w-lg space-y-3 rounded-3xl bg-white p-6"><h3 className="text-xl font-semibold">ساخت Template Draft</h3><input required dir="ltr" className="w-full rounded-xl border p-3" placeholder="group_code" value={form.template_group_code} onChange={e=>setForm({...form,template_group_code:e.target.value,start_permission_code:`workflow_start.${e.target.value}`})}/><input required className="w-full rounded-xl border p-3" placeholder="عنوان فارسی" value={form.name_fa} onChange={e=>setForm({...form,name_fa:e.target.value})}/><textarea className="w-full rounded-xl border p-3" placeholder="توضیحات" value={form.description_fa} onChange={e=>setForm({...form,description_fa:e.target.value})}/><input required dir="ltr" className="w-full rounded-xl border p-3" placeholder="permission code" value={form.start_permission_code} onChange={e=>setForm({...form,start_permission_code:e.target.value})}/><div className="flex gap-2"><button className="rounded-full bg-primary px-5 py-2 text-sand">ساخت</button><button type="button" onClick={()=>setOpen(false)} className="rounded-full border px-5 py-2">انصراف</button></div></form></div>}</div>;
}
