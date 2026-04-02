import { Code2, Database, HardDrive, Layers, Leaf, Server, Zap } from "lucide-react"

export const ServiceIcon = ({ name }: { name: string }) => {
  const sz = 16
  switch (name) {
    case 'apache':   return <Server    size={sz} className="text-orange-400" />
    case 'nginx':    return <Zap       size={sz} className="text-green-400"  />
    case 'mysql':    return <Database  size={sz} className="text-blue-400"   />
    case 'postgres': return <HardDrive size={sz} className="text-sky-400"    />
    case 'mongodb':  return <Leaf      size={sz} className="text-emerald-400"/>
    case 'php':      return <Code2     size={sz} className="text-purple-400" />
    case 'redis':    return <Layers    size={sz} className="text-red-400"    />
    default:         return <Server    size={sz} className="text-gray-400"   />
  }
}