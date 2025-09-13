'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Image from 'next/image';
import { useDrag } from '@use-gesture/react';

const ProductPage = () => {
  const [product, setProduct] = useState<any>(null);
  const [offers, setOffers] = useState([]);
  const [loadingOffers, setLoadingOffers] = useState(false);
  const params = useParams();
  const router = useRouter();
  const { id } = params;

  useEffect(() => {
    if (id) {
      // Fetch product details from the backend
      const fetchProduct = async () => {
        const response = await fetch(`http://localhost:8081/api/product/${id}`);
        const data = await response.json();
        setProduct(data);
      };
      fetchProduct();
    }
  }, [id]);

  const handleShowOffers = async (productId: string) => {
    setLoadingOffers(true);
    const response = await fetch(`http://localhost:8081/api/product/${productId}/offers`);
    const data = await response.json();
    setOffers(data);
    setLoadingOffers(false);
  };

  const bind = useDrag(({ down, movement: [mx], direction: [xDir], velocity: [vx] }) => {
    const trigger = vx > 0.5; // If velocity is high enough, trigger a swipe
    if (!down && trigger && xDir > 0) {
      router.back();
    }
  });

  if (!product) {
    return <div>Loading...</div>;
  }

  return (
    <div {...bind()} className="flex flex-col min-h-screen bg-white dark:bg-gray-800">
      <header className="p-4 border-b border-gray-200 dark:border-gray-700">
        <h1 className="text-2xl font-bold text-center text-gray-900 dark:text-white">{product.title}</h1>
      </header>
      <main className="flex-1 p-4">
        <div className="max-w-4xl mx-auto">
          <div className="flex">
            <div className="w-96 h-96 relative mr-6">
              <Image
                src={product.image_url || '/next.svg'}
                alt={product.title}
                layout="fill"
                objectFit="cover"
                className="rounded-lg"
              />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Specifications</h3>
              <ul className="list-disc list-inside mt-2 text-gray-700 dark:text-gray-300">
                {product.features?.map((feature: string, index: number) => (
                  <li key={index}>{feature}</li>
                ))}
              </ul>
              <div className="mt-6">
                <button
                  onClick={() => handleShowOffers(product.id)}
                  className="px-4 py-2 bg-blue-500 text-white rounded-lg"
                >
                  Show prices and stores
                </button>
              </div>
              {loadingOffers && <p>Loading offers...</p>}
              {offers && offers.length > 0 && (
                <div className="mt-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Offers</h3>
                  <ul className="space-y-2 mt-2">
                    {offers.map((offer: any, index: number) => (
                      <li key={index} className="flex justify-between items-center bg-gray-100 dark:bg-gray-700 p-2 rounded-lg">
                        <a href={offer.link} target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:underline">{offer.merchant}</a>
                        <span className="font-bold text-gray-900 dark:text-white">{offer.price}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          </div>
        </div>
      </main>
      <div className="fixed bottom-4 right-4">
        <button onClick={() => router.back()} className="bg-gray-900 text-white rounded-full p-4 shadow-lg">
          &larr; Back to results
        </button>
      </div>
    </div>
  );
};

export default ProductPage;