import { useState } from 'react'

function App() {
  const [targetUrl, setTargetUrl] = useState('')
  const [shortUrl, setShortUrl] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleShorten = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    
    try {
      // NOTE: Ensure the JSON body matches what your Go API expects!
      const response = await fetch(`${import.meta.env.VITE_API_BASE_URL}/shorten`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: targetUrl }) 
      })
      
      if (!response.ok) throw new Error('Failed to shorten URL')
      
      const data = await response.json()
      setShortUrl(`${import.meta.env.VITE_API_BASE_URL}/${data.code}`)
    } catch (error) {
      console.error("System Error:", error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center font-sans p-4">
      <div className="w-full max-w-md bg-slate-900 p-8 rounded-xl border border-cyan-500/20 shadow-[0_0_15px_rgba(6,182,212,0.15)] relative overflow-hidden">
        
        {/* Minimalist UI Accents */}
        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-cyan-500 to-amber-500"></div>
        
        <h1 className="text-3xl text-slate-100 font-light mb-2 tracking-wide">
          SYS<span className="text-cyan-400 font-bold">LINK</span>
        </h1>
        <p className="text-slate-400 text-sm mb-8">Secure URL Minimization Protocol</p>

        <form onSubmit={handleShorten} className="space-y-6">
          <div>
            <label className="block text-xs font-semibold text-cyan-500 uppercase tracking-widest mb-2">
              Target URL
            </label>
            <input
              type="url"
              required
              value={targetUrl}
              onChange={(e) => setTargetUrl(e.target.value)}
              placeholder="https://..."
              className="w-full bg-slate-950 border border-slate-800 text-slate-200 px-4 py-3 rounded focus:outline-none focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 transition-colors"
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-cyan-950 text-cyan-400 border border-cyan-700/50 hover:bg-cyan-900 hover:border-cyan-400 px-4 py-3 rounded font-medium tracking-wide transition-all disabled:opacity-50"
          >
            {isLoading ? 'PROCESSING...' : 'INITIALIZE SHORTENING'}
          </button>
        </form>

        {shortUrl && (
          <div className="mt-8 p-4 bg-slate-950 border border-amber-500/30 rounded">
            <label className="block text-xs font-semibold text-amber-500 uppercase tracking-widest mb-2">
              Generated Link
            </label>
            <a 
              href={shortUrl} 
              target="_blank" 
              rel="noreferrer"
              className="text-slate-200 break-all hover:text-white transition-colors"
            >
              {shortUrl}
            </a>
          </div>
        )}
      </div>
    </div>
  )
}

export default App