import '@mantine/core/styles.css';
import { MantineProvider } from '@mantine/core';
import { Routes, Route } from 'react-router';

import { MainLayout } from "./layouts/MainLayouts/MainLayout.jsx";
import { ProjectsPage } from './pages/ProjectsPage/ProjectsPage';
import { CalculationPage } from './pages/CalculationPage/CalculationPage';
import { ManualPage } from "./pages/ManualPage/ManualPage"
import { AuthorizationPage } from "./pages/AuthorizationPage/AuthorizationPage";
import { RegistrationPage } from "./pages/RegistrationPage/RegistrationPage";
import { ResetPasswordPage } from "./pages/ResetPasswordPage/ResetPasswordPage";
import { ConfirmCodePage } from "./pages/ConfirmCodePage/ConfirmCodePage";

function App(){
    return(
        <MantineProvider>
            <Routes>
                <Route element={<MainLayout />}>
                    <Route path="/" element={<AuthorizationPage />} />
                    <Route path="/registration" element={<RegistrationPage />} />
                    <Route path="/reset-password" element={<ResetPasswordPage />} />
                    <Route path="/confirm-code" element={<ConfirmCodePage />} />
                    <Route path="/projects" element={<ProjectsPage />}/>
                    <Route path="/calculation" element={<CalculationPage />}/>
                    <Route path="/calculation/:calculationId" element={<CalculationPage />}/>
                    <Route path="/manual" element ={<ManualPage />}/>
                </Route>
            </Routes>
        </MantineProvider>   
    ); 
}

export default App;
