export interface HealthResponse{
    status: string;
    message: string;
}

export async function getHealth(): Promise<HealthResponse>{
    const response = await fetch("http://localhost:8080/health");

    if(!response.ok){
        throw new Error(`API request failed: ${response.status}`);
    }

    return response.json();
}