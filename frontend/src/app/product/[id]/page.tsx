'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { motion, AnimatePresence } from 'framer-motion';
import { useDrag } from '@use-gesture/react';
import { 
  ArrowLeft, 
  Star, 
  ExternalLink, 
  ShoppingCart, 
  Heart,
  Share2,
  ChevronLeft,
  ChevronRight,
  MapPin,
  Truck,
  Shield,
  Clock
} from 'lucide-react';
import Image from 'next/image';
import LoadingSpinner from '@/components/LoadingSpinner';

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

interface Offer {
  merchant: string;
  price: number;
  link: string;
  rating?: number;
  shipping?: string;
  availability?: string;
}

export default function ProductPage() {
  const [product, setProduct] = useState<Product | null>(null);
  const [offers, setOffers] = useState<Offer[]>([]);
  const [loadingProduct, setLoadingProduct] = useState(true);
  const [loadingOffers, setLoadingOffers] = useState(false);
  const [showOffers, setShowOffers] = useState(false);
  const [currentImageIndex, setCurrentImageIndex] = useState(0);
  const [isMinimized, setIsMinimized] = useState(false);
  
  const params = useParams();
  const router = useRouter();
  const { id } = params;

  // Mock additional images for demo
  const mockImages = [
    product?.image_url || '/placeholder.jpg',
    '/placeholder-2.jpg',
    '/placeholder-3.jpg',
    '/placeholder-4.jpg'
  ].filter(Boolean);

  useEffect(() => {
    if (id) {
      fetchProduct();
    }
  }, [id]);

  const fetchProduct = async () => {
    try {
      setLoadingProduct(true);
      // For demo purposes, we'll create a mock product since the backend endpoint might not exist
      const mockProduct: Product = {
        id: id as string,
        title: "Apple iPhone 16 Pro Max, 256GB Desert Titanium",
        category: "Smartphones",
        features: ["256GB Storage", "Desert Titanium", "Pro Camera System", "A18 Pro Chip", "6.9-inch Display"],
        image_url: "https://images.unsplash.com/photo-1592750475338-74b7b21085ab?w=500&h=500&fit=crop",
        price: {
          price_eur: 1299.99,
          price_gbp: "£1,199.00",
          offer_count: "15+ offers"
        }
      };
      setProduct(mockProduct);
    } catch (error) {
      console.error('Failed to fetch product:', error);
    } finally {
      setLoadingProduct(false);
    }
  };

  const handleShowOffers = async () => {
    if (!product?.id) return;
    
    setLoadingOffers(true);
    setShowOffers(true);
    
    try {
      // Mock offers for demo
      const mockOffers: Offer[] = [
        {
          merchant: "Apple Store",
          price: 1299.99,
          link: "https://apple.com",
          rating: 4.8,
          shipping: "Free shipping",
          availability: "In stock"
        },
        {
          merchant: "Amazon",
          price: 1279.99,
          link: "https://amazon.com",
          rating: 4.6,
          shipping: "Free Prime shipping",
          availability: "2-3 days"
        },
        {
          merchant: "Best Buy",
          price: 1299.99,
          link: "https://bestbuy.com",
          rating: 4.5,
          shipping: "$9.99 shipping",
          availability: "Pick up today"
        }
      ];
      setOffers(mockOffers);
    } catch (error) {
      console.error('Failed to fetch offers:', error);
    } finally {
      setLoadingOffers(false);
    }
  };

  // Gesture handling for swipe to close
  const bind = useDrag(({ down, movement: [mx], direction: [xDir], velocity: [vx] }) => {
    const trigger = vx > 0.5;
    if (!down && trigger && xDir > 0) {
      router.back();
    }
  });

  const nextImage = () => {
    setCurrentImageIndex((prev) => (prev + 1) % mockImages.length);
  };

  const prevImage = () => {
    setCurrentImageIndex((prev) => (prev - 1 + mockImages.length) % mockImages.length);
  };

  if (loadingProduct) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
        <LoadingSpinner size="lg" text="Loading product details..." />
      </div>
    );
  }

  if (!product) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">Product not found</h1>
          <button
            onClick={() => router.back()}
            className="px-6 py-3 bg-blue-600 text-white rounded-full hover:bg-blue-700 transition-colors"
          >
            Go Back
          </button>
        </div>
      </div>
    );
  }

  return (
    <div 
      {...bind()} 
      className={`min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900 transition-all duration-300 ${
        isMinimized ? 'transform scale-75 origin-bottom-right rounded-2xl shadow-2xl' : ''
      }`}
    >
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-200 dark:border-gray-700">
        <div className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <motion.button
              onClick={() => router.back()}
              className="flex items-center space-x-2 text-gray-600 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
              whileHover={{ x: -4 }}
              whileTap={{ scale: 0.95 }}
            >
              <ArrowLeft className="w-5 h-5" />
              <span>Back to results</span>
            </motion.button>

            <div className="flex items-center space-x-4">
              <motion.button
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
                className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              >
                <Heart className="w-5 h-5 text-gray-600 dark:text-gray-300" />
              </motion.button>
              
              <motion.button
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
                className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              >
                <Share2 className="w-5 h-5 text-gray-600 dark:text-gray-300" />
              </motion.button>

              <motion.button
                onClick={() => setIsMinimized(!isMinimized)}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
                className="p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                title={isMinimized ? "Expand" : "Minimize"}
              >
                <ChevronRight className={`w-5 h-5 text-gray-600 dark:text-gray-300 transition-transform ${isMinimized ? 'rotate-180' : ''}`} />
              </motion.button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
          {/* Product Images */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="space-y-6"
          >
            {/* Main Image */}
            <div className="relative aspect-square bg-white dark:bg-gray-800 rounded-2xl overflow-hidden shadow-lg">
              <Image
                src={mockImages[currentImageIndex]}
                alt={product.title}
                fill
                className="object-cover"
                priority
              />
              
              {/* Image Navigation */}
              {mockImages.length > 1 && (
                <>
                  <button
                    onClick={prevImage}
                    className="absolute left-4 top-1/2 -translate-y-1/2 p-2 bg-white/80 dark:bg-gray-800/80 rounded-full shadow-lg hover:bg-white dark:hover:bg-gray-800 transition-colors"
                  >
                    <ChevronLeft className="w-5 h-5" />
                  </button>
                  
                  <button
                    onClick={nextImage}
                    className="absolute right-4 top-1/2 -translate-y-1/2 p-2 bg-white/80 dark:bg-gray-800/80 rounded-full shadow-lg hover:bg-white dark:hover:bg-gray-800 transition-colors"
                  >
                    <ChevronRight className="w-5 h-5" />
                  </button>
                </>
              )}

              {/* Image Indicators */}
              {mockImages.length > 1 && (
                <div className="absolute bottom-4 left-1/2 -translate-x-1/2 flex space-x-2">
                  {mockImages.map((_, index) => (
                    <button
                      key={index}
                      onClick={() => setCurrentImageIndex(index)}
                      className={`w-2 h-2 rounded-full transition-colors ${
                        index === currentImageIndex 
                          ? 'bg-blue-600' 
                          : 'bg-white/60 hover:bg-white/80'
                      }`}
                    />
                  ))}
                </div>
              )}
            </div>

            {/* Thumbnail Images */}
            {mockImages.length > 1 && (
              <div className="grid grid-cols-4 gap-4">
                {mockImages.map((image, index) => (
                  <motion.button
                    key={index}
                    onClick={() => setCurrentImageIndex(index)}
                    className={`aspect-square rounded-lg overflow-hidden border-2 transition-all ${
                      index === currentImageIndex 
                        ? 'border-blue-600 shadow-lg' 
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
                    }`}
                    whileHover={{ scale: 1.05 }}
                    whileTap={{ scale: 0.95 }}
                  >
                    <Image
                      src={image}
                      alt={`${product.title} - Image ${index + 1}`}
                      width={100}
                      height={100}
                      className="w-full h-full object-cover"
                    />
                  </motion.button>
                ))}
              </div>
            )}
          </motion.div>

          {/* Product Details */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="space-y-8"
          >
            {/* Product Info */}
            <div>
              {product.category && (
                <span className="inline-block px-3 py-1 bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200 text-sm font-medium rounded-full mb-4">
                  {product.category}
                </span>
              )}
              
              <h1 className="text-3xl md:text-4xl font-bold text-gray-900 dark:text-white mb-4">
                {product.title}
              </h1>

              {/* Rating */}
              <div className="flex items-center space-x-4 mb-6">
                <div className="flex items-center">
                  {[...Array(5)].map((_, i) => (
                    <Star
                      key={i}
                      className={`w-5 h-5 ${
                        i < 4 ? 'text-yellow-400 fill-current' : 'text-gray-300'
                      }`}
                    />
                  ))}
                </div>
                <span className="text-gray-600 dark:text-gray-400">4.2 (1,234 reviews)</span>
              </div>

              {/* Price */}
              {product.price && (
                <div className="mb-8">
                  <div className="flex items-baseline space-x-4">
                    <span className="text-4xl font-bold text-gray-900 dark:text-white">
                      €{product.price.price_eur?.toFixed(2)}
                    </span>
                    {product.price.price_gbp && (
                      <span className="text-xl text-gray-500 line-through">
                        {product.price.price_gbp}
                      </span>
                    )}
                  </div>
                  {product.price.offer_count && (
                    <p className="text-green-600 dark:text-green-400 font-medium mt-2">
                      {product.price.offer_count} available
                    </p>
                  )}
                </div>
              )}
            </div>

            {/* Features */}
            {product.features && product.features.length > 0 && (
              <div>
                <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
                  Key Features
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  {product.features.map((feature, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: index * 0.1 }}
                      className="flex items-center space-x-3 p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
                    >
                      <div className="w-2 h-2 bg-blue-600 rounded-full" />
                      <span className="text-gray-700 dark:text-gray-300">{feature}</span>
                    </motion.div>
                  ))}
                </div>
              </div>
            )}

            {/* Action Buttons */}
            <div className="space-y-4">
              <motion.button
                onClick={handleShowOffers}
                disabled={loadingOffers}
                className="w-full flex items-center justify-center space-x-2 px-8 py-4 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-semibold rounded-xl transition-colors shadow-lg hover:shadow-xl"
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
              >
                {loadingOffers ? (
                  <LoadingSpinner size="sm" />
                ) : (
                  <>
                    <ShoppingCart className="w-5 h-5" />
                    <span>Show Prices & Stores</span>
                  </>
                )}
              </motion.button>

              <div className="grid grid-cols-3 gap-4 text-center">
                <div className="flex flex-col items-center space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <Truck className="w-6 h-6 text-blue-600" />
                  <span className="text-sm text-gray-600 dark:text-gray-300">Free Shipping</span>
                </div>
                <div className="flex flex-col items-center space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <Shield className="w-6 h-6 text-green-600" />
                  <span className="text-sm text-gray-600 dark:text-gray-300">Warranty</span>
                </div>
                <div className="flex flex-col items-center space-y-2 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <Clock className="w-6 h-6 text-purple-600" />
                  <span className="text-sm text-gray-600 dark:text-gray-300">Fast Delivery</span>
                </div>
              </div>
            </div>
          </motion.div>
        </div>

        {/* Offers Section */}
        <AnimatePresence>
          {showOffers && (
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -20 }}
              className="mt-12"
            >
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
                Available Offers
              </h2>
              
              {loadingOffers ? (
                <div className="flex justify-center py-12">
                  <LoadingSpinner size="lg" text="Loading offers..." />
                </div>
              ) : (
                <div className="grid gap-4">
                  {offers.map((offer, index) => (
                    <motion.div
                      key={index}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      transition={{ delay: index * 0.1 }}
                      className="flex items-center justify-between p-6 bg-white dark:bg-gray-800 rounded-xl shadow-sm hover:shadow-md transition-all border border-gray-200 dark:border-gray-700"
                    >
                      <div className="flex items-center space-x-4">
                        <div className="w-12 h-12 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                          <span className="text-white font-bold text-lg">
                            {offer.merchant.charAt(0)}
                          </span>
                        </div>
                        
                        <div>
                          <h3 className="font-semibold text-gray-900 dark:text-white">
                            {offer.merchant}
                          </h3>
                          <div className="flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-400">
                            {offer.rating && (
                              <div className="flex items-center space-x-1">
                                <Star className="w-4 h-4 text-yellow-400 fill-current" />
                                <span>{offer.rating}</span>
                              </div>
                            )}
                            {offer.shipping && (
                              <div className="flex items-center space-x-1">
                                <Truck className="w-4 h-4" />
                                <span>{offer.shipping}</span>
                              </div>
                            )}
                            {offer.availability && (
                              <div className="flex items-center space-x-1">
                                <MapPin className="w-4 h-4" />
                                <span>{offer.availability}</span>
                              </div>
                            )}
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center space-x-4">
                        <div className="text-right">
                          <div className="text-2xl font-bold text-gray-900 dark:text-white">
                            €{offer.price.toFixed(2)}
                          </div>
                        </div>
                        
                        <motion.a
                          href={offer.link}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="flex items-center space-x-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-medium rounded-lg transition-colors"
                          whileHover={{ scale: 1.05 }}
                          whileTap={{ scale: 0.95 }}
                        >
                          <span>Visit Store</span>
                          <ExternalLink className="w-4 h-4" />
                        </motion.a>
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      {/* Minimize Button - Fixed Position */}
      {!isMinimized && (
        <motion.button
          onClick={() => setIsMinimized(true)}
          className="fixed bottom-6 right-6 p-4 bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-full shadow-lg hover:shadow-xl transition-all z-50"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.9 }}
          title="Minimize to corner"
        >
          <ChevronRight className="w-6 h-6" />
        </motion.button>
      )}
    </div>
  );
}