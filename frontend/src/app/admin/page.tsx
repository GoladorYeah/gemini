'use client';

import { useState, useEffect } from 'react';

const AdminPage = () => {
  const [parserStatus, setParserStatus] = useState('idle');
  const [url, setUrl] = useState('');
  const [category, setCategory] = useState('');
  const [geminiApiKeys, setGeminiApiKeys] = useState('');
  const [serpApiKeys, setSerpApiKeys] = useState('');
  const [logs, setLogs] = useState('');
  const [selectedService, setSelectedService] = useState('backend');
  const [products, setProducts] = useState<any[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<any>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [statistics, setStatistics] = useState<any>(null);

  const getParserStatus = async () => {
    const response = await fetch('http://localhost:8081/api/admin/parser/status');
    const data = await response.json();
    setParserStatus(data.status);
  };

  const startParser = async () => {
    await fetch('http://localhost:8081/api/admin/parser/start', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ url, category }),
    });
    getParserStatus();
  };

  const stopParser = async () => {
    await fetch('http://localhost:8081/api/admin/parser/stop', {
      method: 'POST',
    });
    getParserStatus();
  };

  const getApiKeys = async () => {
    const response = await fetch('http://localhost:8081/api/admin/keys');
    const data = await response.json();
    setGeminiApiKeys(data.gemini_api_keys);
    setSerpApiKeys(data.serpapi_api_keys);
  };

  const updateApiKeys = async () => {
    await fetch('http://localhost:8081/api/admin/keys', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ gemini_api_keys: geminiApiKeys, serpapi_api_keys: serpApiKeys }),
    });
    getApiKeys();
  };

  const getLogs = async () => {
    const response = await fetch(`http://localhost:8081/api/admin/logs/${selectedService}`);
    const data = await response.text();
    setLogs(data);
  };

  const getProducts = async () => {
    const response = await fetch('http://localhost:8081/api/admin/products/');
    const data = await response.json();
    setProducts(data);
  };

  const addProduct = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const newProduct = {
      id: formData.get('id') as string,
      title: formData.get('title') as string,
      category: formData.get('category') as string,
      features: (formData.get('features') as string).split(','),
      image_url: formData.get('image_url') as string,
    };
    await fetch('http://localhost:8081/api/admin/products/', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(newProduct),
    });
    getProducts();
  };

  const updateProduct = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const updatedProduct = {
      id: selectedProduct.id,
      title: formData.get('title') as string,
      category: formData.get('category') as string,
      features: (formData.get('features') as string).split(','),
      image_url: formData.get('image_url') as string,
    };
    await fetch(`http://localhost:8081/api/admin/products/${selectedProduct.id}`,
      {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedProduct),
      }
    );
    getProducts();
    setIsEditing(false);
    setSelectedProduct(null);
  };

  const deleteProduct = async (productId: string) => {
    await fetch(`http://localhost:8081/api/admin/products/${productId}`, {
      method: 'DELETE',
    });
    getProducts();
  };

  const getStatistics = async () => {
    const response = await fetch('http://localhost:8081/api/admin/statistics');
    const data = await response.json();
    setStatistics(data);
  };

  useEffect(() => {
    getParserStatus();
    getApiKeys();
    getLogs();
    getProducts();
    getStatistics();
  }, [selectedService]);

  return (
    <div className="flex flex-col min-h-screen bg-white dark:bg-gray-800">
      <header className="p-4 border-b border-gray-200 dark:border-gray-700">
        <h1 className="text-2xl font-bold text-center text-gray-900 dark:text-white">Admin Panel</h1>
      </header>
      <main className="flex-1 p-4">
        <div className="max-w-4xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Parser Management */}
            <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">Parser Management</h2>
              <div className="mt-4">
                <p className="text-gray-700 dark:text-gray-300">Status: <span className="font-bold">{parserStatus}</span></p>
                <div className="mt-4">
                  <input
                    type="text"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    placeholder="URL to parse"
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div className="mt-2">
                  <input
                    type="text"
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    placeholder="Product category"
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div className="mt-4 flex space-x-2">
                  <button onClick={startParser} className="px-4 py-2 bg-blue-500 text-white rounded-lg">Start</button>
                  <button onClick={stopParser} className="px-4 py-2 bg-red-500 text-white rounded-lg">Stop</button>
                  <button onClick={getParserStatus} className="px-4 py-2 bg-gray-500 text-white rounded-lg">Refresh</button>
                </div>
              </div>
            </div>

            {/* API Key Management */}
            <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">API Key Management</h2>
              <div className="mt-4">
                <div className="mt-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Gemini API Keys (comma-separated)</label>
                  <input
                    type="text"
                    value={geminiApiKeys}
                    onChange={(e) => setGeminiApiKeys(e.target.value)}
                    placeholder="..."
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div className="mt-2">
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">SerpApi API Keys (comma-separated)</label>
                  <input
                    type="text"
                    value={serpApiKeys}
                    onChange={(e) => setSerpApiKeys(e.target.value)}
                    placeholder="..."
                    className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
                  />
                </div>
                <div className="mt-4 flex space-x-2">
                  <button onClick={updateApiKeys} className="px-4 py-2 bg-blue-500 text-white rounded-lg">Update Keys</button>
                </div>
              </div>
            </div>

            {/* Log Viewer */}
            <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg col-span-2">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">Log Viewer</h2>
              <div className="mt-4">
                <div className="flex items-center space-x-2">
                  <select value={selectedService} onChange={(e) => setSelectedService(e.target.value)} className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
                    <option value="backend">Backend</option>
                    <option value="parser">Parser</option>
                  </select>
                  <button onClick={getLogs} className="px-4 py-2 bg-gray-500 text-white rounded-lg">Refresh</button>
                </div>
                <pre className="mt-4 bg-black text-white p-4 rounded-lg overflow-x-auto">{logs}</pre>
              </div>
            </div>

            {/* Product Management */}
            <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg col-span-2">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">Product Management</h2>
              <div className="mt-4">
                <form onSubmit={addProduct} className="space-y-2">
                  <input name="id" placeholder="ID" className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                  <input name="title" placeholder="Title" className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                  <input name="category" placeholder="Category" className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                  <input name="features" placeholder="Features (comma-separated)" className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                  <input name="image_url" placeholder="Image URL" className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                  <button type="submit" className="px-4 py-2 bg-green-500 text-white rounded-lg">Add Product</button>
                </form>
                <table className="mt-4 w-full text-left table-auto">
                  <thead>
                    <tr>
                      <th className="px-4 py-2">ID</th>
                      <th className="px-4 py-2">Title</th>
                      <th className="px-4 py-2">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {products.map((product) => (
                      <tr key={product.id}>
                        <td className="border px-4 py-2">{product.id}</td>
                        <td className="border px-4 py-2">{product.title}</td>
                        <td className="border px-4 py-2">
                          <button onClick={() => { setSelectedProduct(product); setIsEditing(true); }} className="px-2 py-1 bg-yellow-500 text-white rounded-lg mr-2">Edit</button>
                          <button onClick={() => deleteProduct(product.id)} className="px-2 py-1 bg-red-500 text-white rounded-lg">Delete</button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Usage Statistics */}
            <div className="bg-gray-100 dark:bg-gray-700 p-4 rounded-lg col-span-2">
              <h2 className="text-xl font-bold text-gray-900 dark:text-white">Usage Statistics</h2>
              {statistics && (
                <div className="mt-4">
                  <p>Total Requests: {statistics.total_requests}</p>
                  <p>Unique Queries: {statistics.unique_queries}</p>
                  <p>Most Popular Query: {statistics.most_popular_query}</p>
                </div>
              )}
            </div>
          </div>
        </div>
        {isEditing && selectedProduct && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 w-full max-w-2xl">
              <form onSubmit={updateProduct} className="space-y-2">
                <input name="title" defaultValue={selectedProduct.title} className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                <input name="category" defaultValue={selectedProduct.category} className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                <input name="features" defaultValue={selectedProduct.features.join(',')} className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                <input name="image_url" defaultValue={selectedProduct.image_url} className="w-full px-4 py-2 border border-gray-300 rounded-lg" />
                <div className="flex justify-end space-x-2">
                  <button type="submit" className="px-4 py-2 bg-blue-500 text-white rounded-lg">Update</button>
                  <button onClick={() => setIsEditing(false)} className="px-4 py-2 bg-gray-500 text-white rounded-lg">Cancel</button>
                </div>
              </form>
            </div>
          </div>
        )}
      </main>
    </div>
  );
};

export default AdminPage;