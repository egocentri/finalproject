import { useState } from 'react'

export default function Calculator() {
  const [expr, setExpr] = useState('')
  const [res, setRes] = useState(null)
  const [err, setErr] = useState('')

  const submit = async (e) => {
    e.preventDefault(); setErr(''); setRes(null)
    const token = localStorage.getItem('token')
    const resp = await fetch('/api/v1/calculate', {
      method:'POST',
      headers:{
        'Content-Type':'application/json',
        'Authorization':`Bearer ${token}`
      },
      body: JSON.stringify({expression: expr})
    })
    if (resp.ok) {
      const data = await resp.json()
      setRes(data.result)
      setExpr('')
    } else {
      if (resp.status===401) { localStorage.removeItem('token'); window.location.reload() }
      else setErr(await resp.text())
    }
  }

  return (
    <div className="max-w-md mx-auto mt-10">
      <h2 className="text-2xl mb-4">Калькулятор</h2>
      <form onSubmit={submit} className="flex gap-2">
        <input value={expr} onChange={e=>setExpr(e.target.value)} placeholder="2+2*2" className="flex-1 p-2 border" required/>
        <button className="px-4 bg-green-600 text-white">=</button>
      </form>
      {res!==null && <div className="mt-4 p-4 border">Результат: <b>{res}</b></div>}
      {err && <div className="text-red-600 mt-4">{err}</div>}
    </div>
  )
}
