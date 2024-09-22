import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}


export const formatXAxis = (tickItem: string) => {
  const date = new Date(tickItem)
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  // const seconds = date.getSeconds().toString().padStart(2, '0');
  return `${hours}:${minutes}`
}

export const formatBytes = (bytes: number): string =>{
  if (bytes < 1024) return `${bytes} B`
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  const formattedSize = (bytes / Math.pow(1024, i)).toFixed(0)
  return `${formattedSize} ${sizes[i]}`
}