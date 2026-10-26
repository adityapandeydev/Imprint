import { useState } from 'react';
import { BookOpen, Star, ChevronDown, Trash2, Check, X } from 'lucide-react';
import type { WishlistItem, ReadingStatus } from '../types/api';
interface WishlistCardProps {
  item: WishlistItem; onUpdateStatus: (id: string, status: ReadingStatus) => void; onUpdatePriority: (id: string, priority: number) => void;
  onUpdateRating: (id: string, rating: number) => void; onUpdateNotes: (id: string, notes: string) => void; onUpdateTags?: (id: string, tags: string[]) => void;
  onDelete: (id: string) => void; onInspectEditions?: (item: WishlistItem) => void;
}
export const STATUS_LABELS: Record<ReadingStatus, string> = { WANT_TO_READ:'Want to read', CURRENTLY_READING:'Reading', FINISHED:'Finished', ABANDONED:'Put aside' };
export function WishlistCard({ item, onUpdateStatus, onUpdatePriority, onUpdateRating, onUpdateNotes, onUpdateTags, onDelete, onInspectEditions }: WishlistCardProps) {
  const [expanded,setExpanded] = useState(false);
  const [notes,setNotes] = useState(item.notes || '');
  const [tag,setTag] = useState('');
  const [confirmDelete,setConfirmDelete] = useState(false);
  const [imageFailed,setImageFailed] = useState(false);
  const title = item.work?.title || 'Untitled book';
  const addTag = () => { const clean = tag.trim(); if (clean && !item.tags.some(t=>t.toLowerCase()===clean.toLowerCase())) onUpdateTags?.(item.id,[...item.tags,clean]); setTag(''); };
  return <article className={`collection-card ${expanded ? 'expanded' : ''}`}>
    <div className="collection-card-main"><button className="collection-cover" onClick={() => onInspectEditions?.(item)} aria-label={`View ${title} editions`}>
      {item.work?.cover_url && !imageFailed ? <img src={item.work.cover_url} alt={title} loading="lazy" onError={()=>setImageFailed(true)} /> : <BookOpen size={25} />}
    </button><div className="collection-copy"><h3><button onClick={()=>onInspectEditions?.(item)}>{title}</button></h3><p>{item.work?.authors?.map(a=>a.name).join(', ') || 'Unknown author'}</p><span className={`reading-status status-${item.status.toLowerCase()}`}>{STATUS_LABELS[item.status]}</span>
    <div className="collection-rating" aria-label={`Rating: ${item.rating || 0} out of 5`}>{[1,2,3,4,5].map(r=><button key={r} aria-label={`Rate ${title} ${r} stars`} onClick={()=>onUpdateRating(item.id,r)}><Star size={15} fill={r <= (item.rating || 0) ? 'currentColor' : 'none'} /></button>)}</div>
    </div><button className="icon-button card-expand" aria-label={`Manage ${title}`} aria-expanded={expanded} onClick={()=>{ if (!expanded) setNotes(item.notes || ''); setExpanded(!expanded); }}><ChevronDown size={18} style={{transform:expanded?'rotate(180deg)':undefined}} /></button></div>
    {!expanded && <div className="collection-card-bottom"><span>{item.edition?.format?.replaceAll('_',' ').toLowerCase() || 'No edition selected'}</span><button onClick={()=>{setNotes(item.notes || '');setExpanded(true);}}>Manage book</button></div>}
    {expanded && <div className="collection-editor">
      <div className="editor-fields"><label>Reading status<select aria-label={`Reading status for ${title}`} value={item.status} onChange={e=>onUpdateStatus(item.id,e.target.value as ReadingStatus)}>{Object.entries(STATUS_LABELS).map(([value,label])=><option value={value} key={value}>{label}</option>)}</select></label><label>Priority<select aria-label={`Priority for ${title}`} value={item.priority} onChange={e=>onUpdatePriority(item.id,Number(e.target.value))}>{[1,2,3,4,5].map(p=><option key={p} value={p}>{p} — {['Low','Later','Normal','Soon','Next up'][p-1]}</option>)}</select></label></div>
      <button className="text-action" onClick={()=>onInspectEditions?.(item)}>{item.edition ? 'Change edition' : 'Choose an edition'} <span aria-hidden="true">→</span></button>
      <div className="shelf-tags">{item.tags.map(t=><span key={t}>{t}<button onClick={()=>onUpdateTags?.(item.id,item.tags.filter(x=>x!==t))} aria-label={`Remove shelf ${t}`}><X size={12} /></button></span>)}</div>
      {onUpdateTags && <form className="tag-form" onSubmit={e=>{e.preventDefault();addTag();}}><input value={tag} onChange={e=>setTag(e.target.value)} aria-label={`Add ${title} to shelf`} placeholder="Add to a shelf…" maxLength={60} /><button type="submit" className="button secondary compact" disabled={!tag.trim()}>Add</button></form>}
      <label className="notes-field">Reading notes<textarea value={notes} onChange={e=>setNotes(e.target.value)} rows={3} placeholder="A thought, a quote, a reason to return…" /></label>
      <div className="editor-footer"><button className="button secondary compact" onClick={()=>{onUpdateNotes(item.id,notes);}}><Check size={15} />Save notes</button><button className="icon-button" aria-label={`Remove ${title} from collection`} onClick={()=>setConfirmDelete(true)}><Trash2 size={16} /></button></div>
      {confirmDelete && <div className="delete-confirm" role="alert"><p>Remove this book and its notes?</p><div><button className="button danger compact" onClick={()=>onDelete(item.id)}>Remove book</button><button className="button secondary compact" onClick={()=>setConfirmDelete(false)}>Keep book</button></div></div>}
    </div>}
  </article>;
}
