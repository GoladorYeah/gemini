'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { useRouter } from 'next/navigation';

export default function Home() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [lang, setLang] = useState('en');
  const [region, setRegion] = useState('US');
  const router = useRouter();

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const userLang = navigator.language || 'en-US';
      const langParts = userLang.split('-');
      setLang(langParts[0]);
      setRegion(langParts[1] || 'US');
    }
  }, []);

  const handleSearch = async () => {
    if (!query) return;
    setLoading(true);
    const response = await fetch('http://localhost:8081/api/search', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ query, lang, region }),
    });
    const data = await response.json();
    setResults(data);
    setLoading(false);
  };

  const handleProductClick = (product: any) => {
    router.push(`/product/${product.id}`);
  };

  return (
    <div className="flex flex-col h-screen bg-white dark:bg-gray-800">
      <header className="p-4 border-b border-gray-200 dark:border-gray-700">
        <h1 className="text-2xl font-bold text-center text-gray-900 dark:text-white">Product Search</h1>
      </header>
      <main className="flex-1 overflow-y-auto p-4">
        <div className="max-w-4xl mx-auto">
          {results.length === 0 && !loading && (
            <div className="flex flex-col items-center justify-center h-full">
              <h2 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">Ask me anything!</h2>
              <p className="text-gray-600 dark:text-gray-400">I can help you find any product you want.</p>
            </div>
          )}
          <div className="space-y-4">
            {results.map((product: any) => (
              <div key={product.id} onClick={() => handleProductClick(product)} className="flex items-center bg-gray-100 dark:bg-gray-700 p-4 rounded-lg cursor-pointer hover:bg-gray-200 dark:hover:bg-gray-600">
                <div className="w-24 h-24 relative mr-4">
                  <Image
                    src={product.image_url || '/next.svg'} // Fallback to a default image
                    alt={product.title}
                    layout="fill"
                    objectFit="cover"
                    className="rounded-lg"
                  />
                </div>
                <h2 className="text-xl font-bold text-gray-900 dark:text-white">{product.title}</h2>
              </div>
            ))}
          </div>
        </div>
      </main>
      <footer className="p-4 bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
        <div className="max-w-4xl mx-auto">
          <div className="flex items-center">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
              placeholder="Search for products..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
            />
            <button
              onClick={handleSearch}
              disabled={loading}
              className="ml-2 px-4 py-2 bg-blue-500 text-white rounded-lg disabled:bg-blue-300"
            >
              {loading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </div>
      </footer>
    </div>
  );
}