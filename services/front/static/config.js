// export const API_BASE_URL = 'http://127.0.0.1:8081/api';
export const API_BASE_URL = 'http://89.108.116.240:8081/api';

export const endpoints = {
  basicProjects: `${API_BASE_URL}/projects`,
  projects: `${API_BASE_URL}/projects/`,
  //getProjects: `${API_BASE_URL}/projects`,
  //createProject: `${API_BASE_URL}/projects`,
  loadFile: `${API_BASE_URL}/projects/`, 
  loadFileDocumentation: `/documentation`, 
  startDocCheck: `/checklist`, 
  remarks: `/remarks`, 
  getRemarks: `${API_BASE_URL}/remarks`,
  remarks_clustered: `/remarks_clustered`, 
  finalReport: `/final_report`, 
};
