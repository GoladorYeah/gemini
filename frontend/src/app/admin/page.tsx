'use client';

import { useState, useEffect } from 'react';
import { motion } from 'framer-motion';
import { 
  Settings, 
  Key, 
  FileText, 
  Package, 
  BarChart3, 
  Play, 
  Square, 
  RefreshCw,
  Eye,
  EyeOff,
  Save,
  Trash2,
  Edit,
  Plus,
  Download,
  Upload
} from 'lucide-react';
import LoadingSpinner from '@/components/LoadingSpinner';

interface Product {
  id: string;
  title: string;
  category: string;
  features: string[];
  image_url?: string;
  google_product_id?: string;
}

interface Statistics {
  total_requests: number;
  unique_queries: number;
  most_popular_query: string;
}

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState('parser');
  const [parserStatus, setParserStatus] = useState('idle');
  const [url, setUrl] = useState('');
  const [category, setCategory] = useState('');
  const [geminiApiKeys, setGeminiApiKeys] = useState('');
  const [serpApiKeys, setSerpApiKeys] = useState('');
  const [showKeys, setShowKeys] = useState(false);
  const [logs, setLogs] = useState('');
  const [selectedService, setSelectedService] = useState('backend');
  const [products, setProducts] = useState<Product[]>([]);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [statistics, setStatistics] = useState<Statistics | null>(null);
  const [loading, setLoading] = useState(false);

  const tabs = [
    { id: 'parser', label: 'Parser Management', icon: Settings },
    { id: 'keys', label: 'API Keys', icon: Key },
    { id: 'logs', label: 'System Logs', icon: FileText },
    { id: 'products', label: 'Products', icon: Package },
    { id: 'stats', label: 'Statistics', icon: BarChart3 }
  ];

  useEffect(() => {
    if (activeTab === 'parser') getParserStatus();
    if (activeTab === 'keys') getApiKeys();
    if (activeTab === 'logs') getLogs();
    if (activeTab === 'products') getProducts();
    if (activeTab === 'stats') getStatistics();
  }, [activeTab, selectedService]);

  const getParserStatus = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/admin/parser/status');
      const data = await response.json();
      setParserStatus(data.status);
    } catch (error) {
      console.error('Failed to get parser status:', error);
    }
  };

  const startParser = async () => {
    try {
      setLoading(true);
      await fetch('http://localhost:8081/api/admin/parser/start', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, category }),
      });
      getParserStatus();
    } catch (error) {
      console.error('Failed to start parser:', error);
    } finally {
      setLoading(false);
    }
  };

  const stopParser = async () => {
    try {
      setLoading(true);
      await fetch('http://localhost:8081/api/admin/parser/stop', { method: 'POST' });
      getParserStatus();
    } catch (error) {
      console.error('Failed to stop parser:', error);
    } finally {
      setLoading(false);
    }
  };

  const getApiKeys = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/admin/keys');
      const data = await response.json();
      setGeminiApiKeys(data.gemini_api_keys || '');
      setSerpApiKeys(data.serpapi_api_keys || '');
    } catch (error) {
      console.error('Failed to get API keys:', error);
    }
  };

  const updateApiKeys = async () => {
    try {
      setLoading(true);
      await fetch('http://localhost:8081/api/admin/keys', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          gemini_api_keys: geminiApiKeys, 
          serpapi_api_keys: serpApiKeys 
        }),
      });
      getApiKeys();
    } catch (error) {
      console.error('Failed to update API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const getLogs = async () => {
    try {
      const response = await fetch(`http://localhost:8081/api/admin/logs/${selectedService}`);
      const data = await response.text();
      setLogs(data);
    } catch (error) {
      console.error('Failed to get logs:', error);
      setLogs('Failed to load logs');
    }
  };

  const getProducts = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/admin/products/');
      const data = await response.json();
      setProducts(data || []);
    } catch (error) {
      console.error('Failed to get products:', error);
      setProducts([]);
    }
  };

  const getStatistics = async () => {
    try {
      const response = await fetch('http://localhost:8081/api/admin/statistics');
      const data = await response.json();
      setStatistics(data);
    } catch (error) {
      console.error('Failed to get statistics:', error);
    }
  };

  const deleteProduct = async (productId: string) => {
    try {
      await fetch(`http://localhost:8081/api/admin/products/${productId}`, {
        method: 'DELETE',
      });
      getProducts();
    } catch (error) {
      console.error('Failed to delete product:', error);
    }
  };

  const renderTabContent = () => {
    switch (activeTab) {
      case 'parser':
        return (
          <div className="space-y-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Parser Control
                </h3>
                <div className="flex items-center space-x-2">
                  <div className={`w-3 h-3 rounded-full ${
                    parserStatus === 'running' ? 'bg-green-500' : 
                    parserStatus === 'stopping' ? 'bg-yellow-500' : 'bg-gray-400'
                  }`} />
                  <span className="text-sm font-medium text-gray-600 dark:text-gray-300">
                    Status: {parserStatus}
                  </span>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    URL to Parse
                  </label>
                  <input
                    type="url"
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    placeholder="https://www.pricerunner.com/..."
                    className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Product Category
                  </label>
                  <input
                    type="text"
                    value={category}
                    onChange={(e) => setCategory(e.target.value)}
                    placeholder="Electronics, Clothing, etc."
                    className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>
              </div>

              <div className="flex space-x-4 mt-6">
                <motion.button
                  onClick={startParser}
                  disabled={loading || parserStatus === 'running'}
                  className="flex items-center space-x-2 px-6 py-3 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition-colors"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <Play className="w-4 h-4" />
                  <span>Start Parser</span>
                </motion.button>

                <motion.button
                  onClick={stopParser}
                  disabled={loading || parserStatus === 'idle'}
                  className="flex items-center space-x-2 px-6 py-3 bg-red-600 hover:bg-red-700 disabled:bg-gray-400 text-white font-medium rounded-lg transition-colors"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <Square className="w-4 h-4" />
                  <span>Stop Parser</span>
                </motion.button>

                <motion.button
                  onClick={getParserStatus}
                  className="flex items-center space-x-2 px-6 py-3 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition-colors"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <RefreshCw className="w-4 h-4" />
                  <span>Refresh</span>
                </motion.button>
              </div>
            </div>
          </div>
        );

      case 'keys':
        return (
          <div className="space-y-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  API Key Management
                </h3>
                <button
                  onClick={() => setShowKeys(!showKeys)}
                  className="flex items-center space-x-2 text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400"
                >
                  {showKeys ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  <span>{showKeys ? 'Hide' : 'Show'} Keys</span>
                </button>
              </div>

              <div className="space-y-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Gemini API Keys (comma-separated)
                  </label>
                  <input
                    type={showKeys ? 'text' : 'password'}
                    value={geminiApiKeys}
                    onChange={(e) => setGeminiApiKeys(e.target.value)}
                    placeholder="key1,key2,key3..."
                    className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    SerpApi API Keys (comma-separated)
                  </label>
                  <input
                    type={showKeys ? 'text' : 'password'}
                    value={serpApiKeys}
                    onChange={(e) => setSerpApiKeys(e.target.value)}
                    placeholder="key1,key2,key3..."
                    className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  />
                </div>

                <motion.button
                  onClick={updateApiKeys}
                  disabled={loading}
                  className="flex items-center space-x-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium rounded-lg transition-colors"
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  {loading ? (
                    <LoadingSpinner size="sm" />
                  ) : (
                    <Save className="w-4 h-4" />
                  )}
                  <span>Update Keys</span>
                </motion.button>
              </div>
            </div>
          </div>
        );

      case 'logs':
        return (
          <div className="space-y-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  System Logs
                </h3>
                <div className="flex items-center space-x-4">
                  <select
                    value={selectedService}
                    onChange={(e) => setSelectedService(e.target.value)}
                    className="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                  >
                    <option value="backend">Backend</option>
                    <option value="parser">Parser</option>
                    <option value="frontend">Frontend</option>
                  </select>
                  <motion.button
                    onClick={getLogs}
                    className="flex items-center space-x-2 px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition-colors"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <RefreshCw className="w-4 h-4" />
                    <span>Refresh</span>
                  </motion.button>
                </div>
              </div>

              <div className="bg-black rounded-lg p-4 overflow-auto max-h-96">
                <pre className="text-green-400 text-sm font-mono whitespace-pre-wrap">
                  {logs || 'No logs available'}
                </pre>
              </div>
            </div>
          </div>
        );

      case 'products':
        return (
          <div className="space-y-6">
            <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
                  Product Management
                </h3>
                <div className="flex space-x-4">
                  <motion.button
                    className="flex items-center space-x-2 px-4 py-2 bg-green-600 hover:bg-green-700 text-white font-medium rounded-lg transition-colors"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Plus className="w-4 h-4" />
                    <span>Add Product</span>
                  </motion.button>
                  <motion.button
                    className="flex items-center space-x-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Upload className="w-4 h-4" />
                    <span>Import</span>
                  </motion.button>
                  <motion.button
                    className="flex items-center space-x-2 px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white font-medium rounded-lg transition-colors"
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Download className="w-4 h-4" />
                    <span>Export</span>
                  </motion.button>
                </div>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-left">
                  <thead>
                    <tr className="border-b border-gray-200 dark:border-gray-700">
                      <th className="pb-3 text-sm font-medium text-gray-500 dark:text-gray-400">ID</th>
                      <th className="pb-3 text-sm font-medium text-gray-500 dark:text-gray-400">Title</th>
                      <th className="pb-3 text-sm font-medium text-gray-500 dark:text-gray-400">Category</th>
                      <th className="pb-3 text-sm font-medium text-gray-500 dark:text-gray-400">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {products.map((product) => (
                      <tr key={product.id}>
                        <td className="py-4 text-sm text-gray-900 dark:text-white">
                          {product.id}
                        </td>
                        <td className="py-4 text-sm text-gray-900 dark:text-white">
                          {product.title}
                        </td>
                        <td className="py-4 text-sm text-gray-600 dark:text-gray-300">
                          {product.category}
                        </td>
                        <td className="py-4">
                          <div className="flex space-x-2">
                            <motion.button
                              onClick={() => {
                                setSelectedProduct(product);
                                setIsEditing(true);
                              }}
                              className="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                              whileHover={{ scale: 1.1 }}
                              whileTap={{ scale: 0.9 }}
                            >
                              <Edit className="w-4 h-4" />
                            </motion.button>
                            <motion.button
                              onClick={() => deleteProduct(product.id)}
                              className="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                              whileHover={{ scale: 1.1 }}
                              whileTap={{ scale: 0.9 }}
                            >
                              <Trash2 className="w-4 h-4" />
                            </motion.button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        );

      case 'stats':
        return (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Total Requests</p>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">
                      {statistics?.total_requests || 0}
                    </p>
                  </div>
                  <BarChart3 className="w-8 h-8 text-blue-600" />
                </div>
              </div>

              <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Unique Queries</p>
                    <p className="text-3xl font-bold text-gray-900 dark:text-white">
                      {statistics?.unique_queries || 0}
                    </p>
                  </div>
                  <Package className="w-8 h-8 text-green-600" />
                </div>
              </div>

              <div className="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Popular Query</p>
                    <p className="text-lg font-semibold text-gray-900 dark:text-white truncate">
                      {statistics?.most_popular_query || 'N/A'}
                    </p>
                  </div>
                  <FileText className="w-8 h-8 text-purple-600" />
                </div>
              </div>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
      {/* Header */}
      <header className="bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              Admin Dashboard
            </h1>
            <a
              href="/"
              className="text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
            >
              ← Back to Search
            </a>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-6 py-8">
        <div className="flex flex-col lg:flex-row gap-8">
          {/* Sidebar */}
          <div className="lg:w-64">
            <nav className="space-y-2">
              {tabs.map((tab) => (
                <motion.button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center space-x-3 px-4 py-3 rounded-lg text-left transition-colors ${
                    activeTab === tab.id
                      ? 'bg-blue-600 text-white'
                      : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'
                  }`}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                >
                  <tab.icon className="w-5 h-5" />
                  <span className="font-medium">{tab.label}</span>
                </motion.button>
              ))}
            </nav>
          </div>

          {/* Main Content */}
          <div className="flex-1">
            <motion.div
              key={activeTab}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3 }}
            >
              {renderTabContent()}
            </motion.div>
          </div>
        </div>
      </div>
    </div>
  );
}