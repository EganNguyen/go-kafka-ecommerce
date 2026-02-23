import { useState, useEffect, useCallback } from 'react'

const API_BASE = '/api'

function App() {
    const [products, setProducts] = useState([])
    const [orders, setOrders] = useState([])
    const [cart, setCart] = useState([])
    const [cartOpen, setCartOpen] = useState(false)
    const [activeTab, setActiveTab] = useState('products')
    const [loading, setLoading] = useState(true)
    const [toast, setToast] = useState(null)
    const [checkingOut, setCheckingOut] = useState(false)

    const showToast = useCallback((message, icon = '✅') => {
        setToast({ message, icon })
        setTimeout(() => setToast(null), 3000)
    }, [])

    // Fetch products
    useEffect(() => {
        setLoading(true)
        fetch(`${API_BASE}/products`)
            .then(res => res.json())
            .then(data => {
                setProducts(data || [])
                setLoading(false)
            })
            .catch(err => {
                console.error('Failed to fetch products:', err)
                setLoading(false)
            })
    }, [])

    // Fetch orders when tab switches
    useEffect(() => {
        if (activeTab !== 'orders') return
        fetch(`${API_BASE}/orders`)
            .then(res => res.json())
            .then(data => setOrders(data || []))
            .catch(err => console.error('Failed to fetch orders:', err))
    }, [activeTab])

    const addToCart = (product) => {
        setCart(prev => {
            const existing = prev.find(item => item.product_id === product.id)
            if (existing) {
                return prev.map(item =>
                    item.product_id === product.id
                        ? { ...item, quantity: item.quantity + 1 }
                        : item
                )
            }
            return [...prev, {
                product_id: product.id,
                name: product.name,
                price: product.price,
                quantity: 1
            }]
        })
        showToast(`Added "${product.name}" to cart`, '🛒')
    }

    const updateCartQty = (productId, delta) => {
        setCart(prev => {
            return prev
                .map(item =>
                    item.product_id === productId
                        ? { ...item, quantity: item.quantity + delta }
                        : item
                )
                .filter(item => item.quantity > 0)
        })
    }

    const cartTotal = cart.reduce((sum, item) => sum + item.price * item.quantity, 0)
    const cartCount = cart.reduce((sum, item) => sum + item.quantity, 0)

    const checkout = async () => {
        if (cart.length === 0) return
        setCheckingOut(true)

        try {
            const res = await fetch(`${API_BASE}/orders`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ items: cart }),
            })

            if (!res.ok) throw new Error('Failed to place order')
            const data = await res.json()

            setCart([])
            setCartOpen(false)
            showToast(`Order placed! ID: ${data.order_id.substring(0, 8)}...`, '🎉')

            // Refresh products (stock may have changed)
            const prodRes = await fetch(`${API_BASE}/products`)
            const prodData = await prodRes.json()
            setProducts(prodData || [])
        } catch (err) {
            showToast('Failed to place order. Please try again.', '❌')
            console.error(err)
        } finally {
            setCheckingOut(false)
        }
    }

    return (
        <div className="app">
            {/* Header */}
            <header className="header">
                <div className="header-brand">
                    <div className="header-logo">W</div>
                    <div>
                        <h1>Watermill Store</h1>
                        <div className="header-subtitle">Event-Driven E-commerce • Go + Kafka</div>
                    </div>
                </div>
                <button className="header-cart" onClick={() => setCartOpen(true)} id="cart-button">
                    🛒 Cart
                    {cartCount > 0 && <span className="cart-badge">{cartCount}</span>}
                </button>
            </header>

            {/* Tabs */}
            <div className="tabs">
                <button
                    className={`tab ${activeTab === 'products' ? 'active' : ''}`}
                    onClick={() => setActiveTab('products')}
                    id="tab-products"
                >
                    Products
                </button>
                <button
                    className={`tab ${activeTab === 'orders' ? 'active' : ''}`}
                    onClick={() => setActiveTab('orders')}
                    id="tab-orders"
                >
                    Orders
                </button>
            </div>

            {/* Products */}
            {activeTab === 'products' && (
                loading ? (
                    <div className="loading">
                        <div className="spinner"></div>
                        Loading products...
                    </div>
                ) : (
                    <div className="products-grid">
                        {products.map(product => (
                            <div key={product.id} className="product-card" id={`product-${product.id}`}>
                                <img
                                    className="product-image"
                                    src={product.image_url}
                                    alt={product.name}
                                    loading="lazy"
                                />
                                <div className="product-body">
                                    <div className="product-category">{product.category}</div>
                                    <h3 className="product-name">{product.name}</h3>
                                    <p className="product-description">{product.description}</p>
                                    <div className="product-footer">
                                        <span className="product-price">${product.price.toFixed(2)}</span>
                                        <span className={`product-stock ${product.stock > 30 ? 'in-stock' : 'low-stock'}`}>
                                            {product.stock > 30 ? `${product.stock} in stock` : `Only ${product.stock} left`}
                                        </span>
                                    </div>
                                    <button
                                        className="add-to-cart-btn"
                                        onClick={() => addToCart(product)}
                                        id={`add-${product.id}`}
                                    >
                                        Add to Cart
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )
            )}

            {/* Orders */}
            {activeTab === 'orders' && (
                <div className="orders-section">
                    <h2>Recent Orders</h2>
                    {orders.length === 0 ? (
                        <div className="loading">No orders yet. Place your first order!</div>
                    ) : (
                        <table className="orders-table">
                            <thead>
                                <tr>
                                    <th>Order ID</th>
                                    <th>Items</th>
                                    <th>Total</th>
                                    <th>Status</th>
                                    <th>Date</th>
                                </tr>
                            </thead>
                            <tbody>
                                {orders.map(order => (
                                    <tr key={order.id}>
                                        <td style={{ fontFamily: 'monospace', fontSize: '0.8rem' }}>
                                            {order.id.substring(0, 8)}...
                                        </td>
                                        <td>
                                            {order.items
                                                ? order.items.map(i => `${i.name} ×${i.quantity}`).join(', ')
                                                : '-'
                                            }
                                        </td>
                                        <td style={{ fontWeight: 600 }}>${order.total_price.toFixed(2)}</td>
                                        <td>
                                            <span className={`status-badge ${order.status}`}>{order.status}</span>
                                        </td>
                                        <td style={{ color: 'var(--text-secondary)', fontSize: '0.82rem' }}>
                                            {new Date(order.created_at).toLocaleString()}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    )}
                </div>
            )}

            {/* Cart Sidebar */}
            {cartOpen && (
                <>
                    <div className="cart-overlay" onClick={() => setCartOpen(false)}></div>
                    <div className="cart-sidebar">
                        <div className="cart-header">
                            <h2>Your Cart ({cartCount})</h2>
                            <button className="cart-close" onClick={() => setCartOpen(false)} id="cart-close">✕</button>
                        </div>

                        <div className="cart-items">
                            {cart.length === 0 ? (
                                <div className="cart-empty">
                                    <div className="cart-empty-icon">🛒</div>
                                    <p>Your cart is empty</p>
                                </div>
                            ) : (
                                cart.map(item => (
                                    <div key={item.product_id} className="cart-item">
                                        <div className="cart-item-info">
                                            <div className="cart-item-name">{item.name}</div>
                                            <div className="cart-item-price">
                                                ${item.price.toFixed(2)} × {item.quantity} = ${(item.price * item.quantity).toFixed(2)}
                                            </div>
                                        </div>
                                        <div className="cart-item-qty">
                                            <button
                                                className="qty-btn"
                                                onClick={() => updateCartQty(item.product_id, -1)}
                                            >
                                                −
                                            </button>
                                            <span>{item.quantity}</span>
                                            <button
                                                className="qty-btn"
                                                onClick={() => updateCartQty(item.product_id, 1)}
                                            >
                                                +
                                            </button>
                                        </div>
                                    </div>
                                ))
                            )}
                        </div>

                        <div className="cart-footer">
                            <div className="cart-total">
                                <span className="cart-total-label">Total</span>
                                <span className="cart-total-value">${cartTotal.toFixed(2)}</span>
                            </div>
                            <button
                                className="checkout-btn"
                                onClick={checkout}
                                disabled={cart.length === 0 || checkingOut}
                                id="checkout-button"
                            >
                                {checkingOut ? 'Placing Order...' : 'Place Order →'}
                            </button>
                        </div>
                    </div>
                </>
            )}

            {/* Toast */}
            {toast && (
                <div className="toast">
                    <span>{toast.icon}</span>
                    {toast.message}
                </div>
            )}
        </div>
    )
}

export default App
