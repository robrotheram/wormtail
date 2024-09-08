export enum  RouterStatus {
    STARTING = "Starting",
	RUNNING  = "Running",
	STOPPED  = "Stopped",
}

export interface ProxyStats {
    Sent:     number;
	Received: number;
}

export interface Route {
    Id: string;
    Name: string;
    Host:   string;
	Port:   number;
	Status: RouterStatus;
	Stats:  ProxyStats;
    Listen: number;
}

export interface TimeSeries {
    Data: Route[];
}


const baseURL = "http://localhost:8080/api"
export const getRoutes = async():Promise<Route[]> => {
    const response = await fetch(`${baseURL}/routes`)
    return await response.json()
}

export const startRoute = async(name:string) => {
    await fetch(`${baseURL}/routes/${name}/start`, {method: "POST"});
}

export const stopRoute = async(name:string) => {
    await fetch(`${baseURL}/routes/${name}/stop`, {method: "POST"});
}