'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { Sparkles, TrendingUp, Zap, Globe } from 'lucide-react';
import SearchInput from '@/components/SearchInput';
import ProductCard from '@/components/ProductCard';
import LoadingSpinner from '@/components/LoadingSpinner';
import { detectLanguage, detectRegion } from '@/lib/utils';

interface Product {
  id: string;
  title: string;
  category?: string;
  features?: string[];
  google_product_id?: string;
  image_url?: string;
  price?: {
    price_gbp?: string;
    price_eur?: number;
    offer_count?: string;
  };
}

export default function Home() {
  const [results, setResults] = useState<Product[]>([]);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);
  const router = useRouter();

  const handleSearch = async (query: string) => {
    if (!query.trim()) return;
    
    setLoading(true);
    setHasSearched(true);
    
    try {
      // Auto-detect language and region
      const lang = detectLanguage(query);
      const region = detectRegion(lang);
      
      const response = await fetch('http://localhost:8081/api/search', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query, lang, region }),
      });
      
      if (!response.ok) {
        throw new Error('Search failed');
      }
      
      const data = await response.json();
      setResults(data || []);
    } catch (error) {
      console.error('Search error:', error);
      setResults([]);
    } finally {
      setLoading(false);
    }
  };

  const handleProductClick = (product: Product) => {
    router.push(`/product/${product.id}`);
  };

  const trendingSearches = [
    "iPhone 15 Pro Max",
    "Samsung Galaxy S24 Ultra",
    "MacBook Air M3",
    "AirPods Pro 2",
    "PlayStation 5",
    "Nintendo Switch OLED"
  ];

  const features = [
    {
      icon: Sparkles,
      title: "AI-Powered Search",
      description: "Advanced AI understands your queries in any language"
    },
    {
      icon: TrendingUp,
      title: "Real-time Prices",
      description: "Compare prices across thousands of merchants instantly"
    },
    {
      icon: Zap,
      title: "Lightning Fast",
      description: "Get results in milliseconds with our optimized search"
    },
    {
      icon: Globe,
      title: "Global Coverage",
      description: "Search products from merchants worldwide"
    }
  ];

  return (
    <div className="min-h-screen flex flex-col">
      {/* Header */}
      <header className="relative z-20 p-6">
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="max-w-7xl mx-auto"
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                <Sparkles className="w-5 h-5 text-white" />
              </div>
              <span className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                Gemini
              </span>
            </div>
            
            <nav className="hidden md:flex items-center space-x-6">
              <a href="#" className="text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                Products
              </a>
              <a href="/admin" className="text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                Admin
              </a>
              <a href="#" className="text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                About
              </a>
            </nav>
          </div>
        </motion.div>
      </header>

      {/* Main Content */}
      <main className="flex-1 px-6">
        <div className="max-w-7xl mx-auto">
          <AnimatePresence mode="wait">
            {!hasSearched ? (
              /* Welcome Screen */
              <motion.div
                key="welcome"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="flex flex-col items-center justify-center min-h-[70vh] text-center"
              >
                <motion.div
                  initial={{ opacity: 0, y: 30 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6 }}
                  className="mb-8"
                >
                  <h1 className="text-6xl md:text-7xl font-bold mb-6">
                    <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                      Find anything
                    </span>
                    <br />
                    <span className="text-gray-900 dark:text-white">
                      you need
                    </span>
                  </h1>
                  <p className="text-xl text-gray-600 dark:text-gray-300 max-w-2xl mx-auto leading-relaxed">
                    Search millions of products, compare prices instantly, and discover the best deals 
                    with our AI-powered product search engine.
                  </p>
                </motion.div>

                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: 0.2 }}
                  className="w-full max-w-4xl mb-12"
                >
                  <SearchInput 
                    onSearch={handleSearch} 
                    loading={loading}
                    placeholder="Search for any product... iPhone, MacBook, sneakers, etc."
                  />
                </motion.div>

                {/* Trending Searches */}
                <motion.div
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: 0.4 }}
                  className="mb-16"
                >
                  <p className="text-gray-500 dark:text-gray-400 mb-4">Trending searches:</p>
                  <div className="flex flex-wrap justify-center gap-3">
                    {trendingSearches.map((search, index) => (
                      <motion.button
                        key={search}
                        initial={{ opacity: 0, scale: 0.8 }}
                        animate={{ opacity: 1, scale: 1 }}
                        transition={{ delay: 0.5 + index * 0.1 }}
                        whileHover={{ scale: 1.05 }}
                        whileTap={{ scale: 0.95 }}
                        onClick={() => handleSearch(search)}
                        className="px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-full text-gray-700 dark:text-gray-300 hover:border-blue-300 dark:hover:border-blue-600 hover:text-blue-600 dark:hover:text-blue-400 transition-all duration-200 shadow-sm hover:shadow-md"
                      >
                        {search}
                      </motion.button>
                    ))}
                  </div>
                </motion.div>

                {/* Features */}
                <motion.div
                  initial={{ opacity: 0, y: 30 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.6, delay: 0.6 }}
                  className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 w-full max-w-6xl"
                >
                  {features.map((feature, index) => (
                    <motion.div
                      key={feature.title}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.7 + index * 0.1 }}
                      className="text-center p-6 rounded-2xl bg-white/50 dark:bg-gray-800/50 backdrop-blur-sm border border-gray-200/50 dark:border-gray-700/50"
                    >
                      <div className="w-12 h-12 bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <feature.icon className="w-6 h-6 text-white" />
                      </div>
                      <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
                        {feature.title}
                      </h3>
                      <p className="text-gray-600 dark:text-gray-300 text-sm">
                        {feature.description}
                      </p>
                    </motion.div>
                  ))}
                </motion.div>
              </motion.div>
            ) : (
              /* Search Results */
              <motion.div
                key="results"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="py-8"
              >
                {/* Search Header */}
                <div className="mb-8">
                  <SearchInput 
                    onSearch={handleSearch} 
                    loading={loading}
                    className="mb-6"
                  />
                  
                  {!loading && results.length > 0 && (
                    <motion.p
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="text-gray-600 dark:text-gray-400"
                    >
                      Found {results.length} products
                    </motion.p>
                  )}
                </div>

                {/* Loading State */}
                {loading && (
                  <div className="flex justify-center py-20">
                    <LoadingSpinner size="lg" text="Searching products..." />
                  </div>
                )}

                {/* Results Grid */}
                {!loading && results.length > 0 && (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {results.map((product, index) => (
                      <ProductCard
                        key={product.id}
                        product={product}
                        onClick={handleProductClick}
                        index={index}
                      />
                    ))}
                  </div>
                )}

                {/* No Results */}
                {!loading && hasSearched && results.length === 0 && (
                  <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    className="text-center py-20"
                  >
                    <div className="w-16 h-16 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center mx-auto mb-4">
                      <Sparkles className="w-8 h-8 text-gray-400" />
                    </div>
                    <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
                      No products found
                    </h3>
                    <p className="text-gray-600 dark:text-gray-400 mb-6">
                      Try adjusting your search terms or browse our trending products.
                    </p>
                    <button
                      onClick={() => setHasSearched(false)}
                      className="px-6 py-3 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
                    >
                      Start New Search
                    </button>
                  </motion.div>
                )}
              </motion.div>
            )}
          </AnimatePresence>
        </div>
      </main>
    </div>
  );
}