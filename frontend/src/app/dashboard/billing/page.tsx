'use client'

import { CreditCard, Download, Check } from 'lucide-react'

export default function BillingPage() {
  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Billing</h1>
        <p className="text-zinc-400">Manage your subscription and payment methods</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Current Plan */}
        <div className="lg:col-span-2 space-y-6">
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <div className="flex items-center justify-between mb-6">
              <div>
                <p className="text-sm text-zinc-400 mb-1">Current Plan</p>
                <h2 className="text-2xl font-bold">Pro</h2>
              </div>
              <div className="text-right">
                <p className="text-3xl font-bold">$29</p>
                <p className="text-zinc-400">/month</p>
              </div>
            </div>

            <div className="p-4 bg-amber-500/10 border border-amber-500/20 rounded-lg mb-6">
              <p className="text-green-400">✓ Your subscription is active</p>
            </div>

            <div className="space-y-3">
              <p className="font-medium mb-3">What's included:</p>
              {[
                '1 AI Assistant deployment',
                'Custom subdomain',
                'Telegram integration',
                '24/7 support',
                'Automatic updates'
              ].map((feature) => (
                <div key={feature} className="flex items-center gap-3 text-zinc-300">
                  <Check className="w-5 h-5 text-amber-500" />
                  {feature}
                </div>
              ))}
            </div>
          </div>

          {/* Payment Method */}
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <h2 className="text-xl font-semibold mb-6">Payment Method</h2>
            
            <div className="flex items-center gap-4 p-4 bg-zinc-800/50 rounded-lg">
              <div className="w-12 h-12 bg-blue-500/10 rounded-lg flex items-center justify-center">
                <CreditCard className="w-6 h-6 text-blue-500" />
              </div>
              <div className="flex-1">
                <p className="font-medium">•••• •••• •••• 4242</p>
                <p className="text-sm text-zinc-400">Expires 12/25</p>
              </div>
              <button className="px-4 py-2 bg-zinc-800 hover:bg-zinc-700 rounded-lg text-sm transition-colors">
                Update
              </button>
            </div>
          </div>
        </div>

        {/* Invoices */}
        <div>
          <div className="bg-zinc-900/50 border border-zinc-800 rounded-xl p-6">
            <h2 className="text-xl font-semibold mb-6">Invoices</h2>
            
            <div className="space-y-4">
              {[
                { date: 'Feb 19, 2026', amount: '$29.00', status: 'Paid' },
                { date: 'Jan 19, 2026', amount: '$29.00', status: 'Paid' },
                { date: 'Dec 19, 2025', amount: '$29.00', status: 'Paid' },
              ].map((invoice, i) => (
                <div key={i} className="flex items-center justify-between p-4 bg-zinc-800/50 rounded-lg">
                  <div>
                    <p className="font-medium">{invoice.date}</p>
                    <p className="text-sm text-zinc-400">{invoice.amount}</p>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="px-2 py-1 bg-amber-500/10 text-green-400 text-xs rounded-full">
                      {invoice.status}
                    </span>
                    <button className="p-2 hover:bg-zinc-700 rounded-lg transition-colors">
                      <Download className="w-4 h-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
