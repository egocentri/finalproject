import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

export default function Register() {
  const [login, setLogin] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')
  const navigate = useNavigate()

  const submit = async (e) => {
    e.preventDefault(); setErr('')
    const res = await fetch('/api/v1/register', {
      method:'POST', headers:{'Content-Type':'application/json'},
      body: JSON.stringify({login,password})
    })
    if (res.ok) navigate('/login')
    else setErr(await res.text())
  }

  return (
    <div className="max-w-sm mx-auto mt-10">
      <h2 className="text-2xl mb-4">Регистрация</h2>
      <form onSubmit={submit} className="space-y-4">
        <input value={login} onChange={e=>setLogin(e.target.value)} placeholder="Логин" className="w-full p-2 border" required/>
        <input type="password" value={password} onChange={e=>setPassword(e.target.value)} placeholder="Пароль" className="w-full p-2 border" required/>
        <button className="w-full bg-blue-600 text-white py-2">Зарегистрироваться</button>
      </form>
      {err && <div className="text-red-600 mt-2">{err}</div>}
    </div>
  )
}
