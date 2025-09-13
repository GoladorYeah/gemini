'use client';

import { useState, useRef, useEffect } from 'react';
import { Search, Loader2, Mic, Camera } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

interface SearchInputProps {
  onSearch: (query: string) => void;
  loading?: boolean;
  placeholder?: string;
  className?: string;
}

export default function SearchInput({ 
  onSearch, 
  loading = false, 
  placeholder = "Search for any product...",
  className 
}: SearchInputProps) {
  const [query, setQuery] = useState('');
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    // Auto-focus on mount
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim() && !loading) {
      onSearch(query.trim());
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSubmit(e);
    }
  };

  return (
    <motion.div
      className={cn(
        "relative w-full max-w-4xl mx-auto",
        className
      )}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <form onSubmit={handleSubmit} className="relative">
        <div className={cn(
          "relative flex items-center bg-white dark:bg-gray-800 rounded-full border-2 transition-all duration-200 shadow-lg",
          isFocused 
            ? "border-blue-500 shadow-xl shadow-blue-500/20" 
            : "border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600"
        )}>
          <div className="flex items-center pl-6 pr-3">
            <Search className="w-5 h-5 text-gray-400" />
          </div>
          
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setIsFocused(true)}
            onBlur={() => setIsFocused(false)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={loading}
            className="flex-1 py-4 px-2 bg-transparent text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none text-lg"
          />

          <div className="flex items-center space-x-2 pr-2">
            {/* Voice Search Button */}
            <motion.button
              type="button"
              className="p-3 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              title="Voice search"
            >
              <Mic className="w-5 h-5 text-gray-500 hover:text-blue-500" />
            </motion.button>

            {/* Image Search Button */}
            <motion.button
              type="button"
              className="p-3 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              title="Search by image"
            >
              <Camera className="w-5 h-5 text-gray-500 hover:text-blue-500" />
            </motion.button>

            {/* Search Button */}
            <motion.button
              type="submit"
              disabled={!query.trim() || loading}
              className={cn(
                "flex items-center justify-center px-6 py-3 rounded-full font-medium transition-all duration-200",
                query.trim() && !loading
                  ? "bg-blue-600 hover:bg-blue-700 text-white shadow-lg hover:shadow-xl"
                  : "bg-gray-100 dark:bg-gray-700 text-gray-400 cursor-not-allowed"
              )}
              whileHover={query.trim() && !loading ? { scale: 1.02 } : {}}
              whileTap={query.trim() && !loading ? { scale: 0.98 } : {}}
            >
              <AnimatePresence mode="wait">
                {loading ? (
                  <motion.div
                    key="loading"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                  >
                    <Loader2 className="w-5 h-5 animate-spin" />
                  </motion.div>
                ) : (
                  <motion.span
                    key="search"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                  >
                    Search
                  </motion.span>
                )}
              </AnimatePresence>
            </motion.button>
          </div>
        </div>
      </form>

      {/* Search Suggestions */}
      <AnimatePresence>
        {isFocused && query.length > 0 && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="absolute top-full mt-2 w-full bg-white dark:bg-gray-800 rounded-2xl shadow-xl border border-gray-200 dark:border-gray-700 overflow-hidden z-50"
          >
            <div className="p-4">
              <div className="text-sm text-gray-500 dark:text-gray-400 mb-2">Recent searches</div>
              <div className="space-y-2">
                {['iPhone 15 Pro', 'Samsung Galaxy S24', 'MacBook Air M3'].map((suggestion, index) => (
                  <motion.button
                    key={suggestion}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: index * 0.05 }}
                    className="flex items-center w-full p-2 text-left hover:bg-gray-50 dark:hover:bg-gray-700 rounded-lg transition-colors"
                    onClick={() => {
                      setQuery(suggestion);
                      onSearch(suggestion);
                    }}
                  >
                    <Search className="w-4 h-4 text-gray-400 mr-3" />
                    <span className="text-gray-700 dark:text-gray-300">{suggestion}</span>
                  </motion.button>
                ))}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  );
}