import { Button } from "./components/ui/button"
import { SheetTrigger, SheetContent, Sheet } from "./components/ui/sheet"
import { LayoutGridIcon, MenuIcon, NetworkIcon, SettingsIcon } from "./Icons"
import { Link } from "@tanstack/react-router"
import React from "react"

interface LinksProps {
  to: string;
  label: string;
  icon: (props: React.SVGProps<SVGSVGElement>) => JSX.Element
  className: string
}
const Links = [
  {
    to: "/routes",
    label: "Routes",
    icon: LayoutGridIcon,
    className: "h-5 w-5"
  },
  {
    to: "/settings",
    label: "Settings",
    icon: SettingsIcon,
    className: "h-5 w-5"
  }
] as LinksProps[]



export const HeaderNav = () => {
  return <header className="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:static sm:h-auto sm:border-0 sm:bg-transparent sm:px-6">
    <Sheet>
      <SheetTrigger asChild>
        <Button size="icon" variant="outline" className="sm:hidden">
          <MenuIcon className="h-5 w-5" />
          <span className="sr-only">Toggle Menu</span>
        </Button>
      </SheetTrigger>
      <SheetContent side="left" className="sm:max-w-xs">
        <nav className="grid gap-6 text-lg font-medium">
          <Link to="/" className="group flex h-10 w-10 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:text-base">
            <NetworkIcon className="h-5 w-5 transition-all group-hover:scale-110" />
          </Link>
          {Links.map((link: LinksProps) => <Link key={link.to} to={link.to} className="flex items-center gap-4 px-2.5 text-foreground">
            {link.icon({ className: link.className })}
            <span className="sr-only">{link.label}</span>
            {link.label}
          </Link>)}
        </nav>
      </SheetContent>
    </Sheet>
  </header>
}
export const SideNav = () => {
  return <aside className="fixed inset-y-0 left-0 z-10 hidden w-14 flex-col border-r bg-background sm:flex">
    <nav className="flex flex-col items-center gap-4 px-2 sm:py-5">
      <Link to="/" className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:h-8 md:w-8 md:text-base">
          <NetworkIcon className="h-4 w-4 transition-all group-hover:scale-110" />
          <span className="sr-only">Load Balancer</span>
      </Link>
      {Links.map((link: LinksProps) => 
        <Link key={link.to} to={link.to} className="group flex h-9 w-9 shrink-0 items-center justify-center gap-2 rounded-full bg-primary text-lg font-semibold text-primary-foreground md:h-8 md:w-8 md:text-base">
        {link.icon({ className: link.className })}
        <span className="sr-only">{link.label}</span>
      </Link>)}
    </nav>
  </aside>
}