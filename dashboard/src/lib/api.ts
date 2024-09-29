export const isDev = !!import.meta.env.DEV;

export enum RouterStatus {
    STARTING = "Starting",
    RUNNING = "Running",
    STOPPED = "Stopped",
}

export enum RouterType {
    HTTP = "http",
    TCP = "tcp",
    UDP = "udp",
}


export interface Route {
    Id: string;
    Name: string;
    Enabled: boolean;
    Type: RouterType;
    Port: number;
    Status: RouterStatus;
    Machine: Machine;
    Stats: TimeSeries;
}

export interface Machine {
    Address: string
    Port: number
}


export interface Tailsale {
    AuthKey: string
    Hostname: string
}

export interface ProxyStats {
    Sent: number;
    Received: number;
}

export interface TimeSeriesPoint {
    Timestamp: Date
    Value: ProxyStats
}
export interface TimeSeries {
    Points: TimeSeriesPoint[]
    Total: ProxyStats
}


export interface Dashboard {
    Enabled: boolean
    Username: string
    Password: string
}

export interface Login {
    username: string
    password: string
}

export interface LoginToken {
    authorization_token	: string
}



export const defaultRoute = {
    Id: "",
    Name: "",
    Enabled: true,
    Type: RouterType.HTTP,
    Machine: {
        Address: "127.0.0.1",
        Port: 0
    },
    Port: 0,
    Status: RouterStatus.STOPPED,
    Stats: {
        Points: [],
        Total: {
            Sent: 0,
            Received: 0,
        }
    }
} as Route

export const token = {
    set: (newToken: string) => sessionStorage.setItem('token', newToken),
    get: () => sessionStorage.getItem('token'),
    remove: () => sessionStorage.removeItem('token')
}

let BASE_URL = ""
if (isDev){
    BASE_URL="http://localhost:8081"
}
const API_URL = `${BASE_URL}/api`
const AUTH_URL = `${BASE_URL}/auth`

const getAuth = () => {
    return {
        Authorization: token.get() || ""
    } as Record<string, string>
}




export const login = async (login: Login): Promise<LoginToken> => {
    const response = await fetch(`${AUTH_URL}/login`, {
        method: "post",
        body: JSON.stringify(login)
    })
    if (response.status != 200) throw Error("UnAthorisied")
    return response.json()
}


export const getRoutes = async (): Promise<Route[]> => {
    const response = await fetch(`${API_URL}/routes`, {
        headers: getAuth(),
    })
    return response.json()
}

export const createRoute = async (route: Route): Promise<Route> => {
    const response = await fetch(`${API_URL}/routes`, {
        headers: getAuth(),
        method: "post",
        body: JSON.stringify(route)
    })
    return response.json()
}

export const getRoute = async (name: string): Promise<Route> => {
    const response = await fetch(`${API_URL}/routes/${name}`, {
        headers: getAuth()
    })
    return response.json()
}


export const updateRoute = async (route: Route): Promise<Route> => {
    const response = await fetch(`${API_URL}/routes/${route.Id}`, {
        headers: getAuth(),
        method: "put",
        body: JSON.stringify(route)
    })
    return response.json()
}

export const deleteRoute = async (route: Route): Promise<Response> => {
    return fetch(`${API_URL}/routes/${route.Id}`, {
        headers: getAuth(),
        method: "delete"
    })
}

export const startRoute = async (name: string) => {
    await fetch(`${API_URL}/routes/${name}/start`, {
        headers: getAuth(),
        method: "POST"
    });
}

export const stopRoute = async (name: string) => {
    await fetch(`${API_URL}/routes/${name}/stop`, {
        headers: getAuth(),
        method: "POST"
    });
}

export const getTSConfig = async (): Promise<Tailsale> => {
    const response = await fetch(`${API_URL}/settings/tailscale`, {
        headers: getAuth(),
    })
    return response.json()
}
export const updateTSConfig = async (config: Tailsale): Promise<Tailsale> => {
    const response = await fetch(`${API_URL}/settings/tailscale`, {
        headers: getAuth(),
        method: "post",
        body: JSON.stringify(config)
    })
    return response.json()
}


export const getDashboardConfig = async (): Promise<Dashboard> => {
    const response = await fetch(`${API_URL}/settings/dashboard`, {
        headers: getAuth(),
    })
    return response.json()
}
export const updateDashboardConfig = async (config: Dashboard): Promise<Dashboard> => {
    const response = await fetch(`${API_URL}/settings/dashboard`, {
        headers: getAuth(),
        method: "post",
        body: JSON.stringify(config)
    })
    return response.json()
}