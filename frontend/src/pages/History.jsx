import { useEffect, useState } from 'react'

export default function History() {
  const [items, setItems] = useState([])
  const [err, setErr] = useState('')

  useEffect(() => {
    const token = localStorage.getItem('token')
    fetch('/api/v1/expressions', { headers: {'Authorization':`Bearer ${token}`} })
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(setItems)
      .catch(e => {
        if (e===401) { localStorage.removeItem('token'); window.location.reload() }
        else setErr('Ошибка загрузки')
      })
  }, [])

  return (
    <div className="max-w-md mx-auto mt-10">
      <h2 className="text-2xl mb-4">История</h2>
      {err && <div className="text-red-600 mb-4">{err}</div>}
      {items.length > 0 ? (
        <ul className="space-y-2">
          {items.map(i => (
            <li key={i.id} className="p-2 border">
              {i.expression} = <b>{i.result}</b>
            </li>
          ))}
        </ul>
      ) : <div>Пусто</div>}
    </div>
  )
}
