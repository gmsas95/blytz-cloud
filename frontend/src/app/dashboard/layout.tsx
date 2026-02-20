import { DashboardSidebar } from '@/components/dashboard-sidebar'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="min-h-screen bg-black text-white flex">
      <DashboardSidebar />
      <main className="flex-1 ml-72">
        {children}
      </main>
    </div>
  )
}
