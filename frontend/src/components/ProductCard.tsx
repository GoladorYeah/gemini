'use client';

import { motion } from 'framer-motion';
import { Star, ExternalLink, ShoppingCart } from 'lucide-react';
import Image from 'next/image';
import { cn } from '@/lib/utils';

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

interface ProductCardProps {
  product: Product;
  onClick: (product: Product) => void;
  index?: number;
}

export default function ProductCard({ product, onClick, index = 0 }: ProductCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ 
        duration: 0.3, 
        delay: index * 0.1,
        ease: "easeOut"
      }}
      whileHover={{ 
        y: -4,
        transition: { duration: 0.2 }
      }}
      className="group cursor-pointer"
      onClick={() => onClick(product)}
    >
      <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-sm hover:shadow-xl transition-all duration-300 border border-gray-100 dark:border-gray-700 overflow-hidden">
        {/* Image Container */}
        <div className="relative aspect-square bg-gray-50 dark:bg-gray-900 overflow-hidden">
          {product.image_url ? (
            <Image
              src={product.image_url}
              alt={product.title}
              fill
              className="object-cover group-hover:scale-105 transition-transform duration-300"
              sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
            />
          ) : (
            <div className="flex items-center justify-center h-full">
              <div className="w-16 h-16 bg-gray-200 dark:bg-gray-700 rounded-full flex items-center justify-center">
                <ShoppingCart className="w-8 h-8 text-gray-400" />
              </div>
            </div>
          )}
          
          {/* Category Badge */}
          {product.category && (
            <div className="absolute top-3 left-3">
              <span className="px-2 py-1 bg-blue-600 text-white text-xs font-medium rounded-full">
                {product.category}
              </span>
            </div>
          )}

          {/* Price Badge */}
          {product.price?.price_eur && (
            <div className="absolute top-3 right-3">
              <span className="px-2 py-1 bg-green-600 text-white text-sm font-semibold rounded-full">
                €{product.price.price_eur.toFixed(2)}
              </span>
            </div>
          )}
        </div>

        {/* Content */}
        <div className="p-4">
          <h3 className="font-semibold text-gray-900 dark:text-white text-lg mb-2 line-clamp-2 group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
            {product.title}
          </h3>

          {/* Features */}
          {product.features && product.features.length > 0 && (
            <div className="flex flex-wrap gap-1 mb-3">
              {product.features.slice(0, 3).map((feature, idx) => (
                <span
                  key={idx}
                  className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 text-xs rounded-md"
                >
                  {feature}
                </span>
              ))}
              {product.features.length > 3 && (
                <span className="px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400 text-xs rounded-md">
                  +{product.features.length - 3} more
                </span>
              )}
            </div>
          )}

          {/* Rating and Reviews (Mock data) */}
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-1">
              <div className="flex items-center">
                {[...Array(5)].map((_, i) => (
                  <Star
                    key={i}
                    className={cn(
                      "w-4 h-4",
                      i < 4 ? "text-yellow-400 fill-current" : "text-gray-300"
                    )}
                  />
                ))}
              </div>
              <span className="text-sm text-gray-600 dark:text-gray-400 ml-1">
                4.2 (1.2k)
              </span>
            </div>

            <motion.div
              className="opacity-0 group-hover:opacity-100 transition-opacity"
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.9 }}
            >
              <ExternalLink className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </motion.div>
          </div>
        </div>
      </div>
    </motion.div>
  );
}