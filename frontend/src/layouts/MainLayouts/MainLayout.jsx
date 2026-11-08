import { AppShell, Text, Stack, Group } from "@mantine/core";
import { useEffect, useState } from "react";
import { Outlet, useLocation, useNavigate } from "react-router";
import { logoutUser } from "../../api/auth";
import avatar from "../../assets/Avatar.svg";
import logo from "../../assets/logo.svg";
import classes from "./MainLayout.module.css";

const AUTH_PATHS = new Set([
  "/",
  "/registration",
  "/reset-password",
  "/confirm-code",
]);

export function MainLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [isUserMenuOpened, setIsUserMenuOpened] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  const showUserButton = !AUTH_PATHS.has(location.pathname);

  useEffect(() => {
    setIsUserMenuOpened(false);
  }, [location.pathname]);

  async function handleLogout(){
    try{
      setIsLoggingOut(true);
      await logoutUser();
    }catch(error){
      console.error(error);
    }finally{
      setIsLoggingOut(false);
      setIsUserMenuOpened(false);
      navigate("/", { replace: true });
    }
  }

  return (
    <AppShell header={{ height: 80 }} className={classes.pageFrame} >
      <AppShell.Header className={classes.header} withBorder={false}>
        <Group justify="space-between" className={classes.HeaderContent}>
          <Group gap={10} className={classes.Brand}>
            <img src={logo} alt="Аккредит" className={classes.Logo} />
            <Stack gap={0}>
              <Text className={classes.sizeText} fw={700} fz={18}>Аккредит</Text>
              <Text className={classes.Text}>Проверка критерия средней заработной платы</Text>
            </Stack>
          </Group>
          {showUserButton && (
            <div className={classes.UserMenu}>
              <button
                type="button"
                className={classes.UserButton}
                aria-label="Профиль"
                aria-expanded={isUserMenuOpened}
                onClick={() => setIsUserMenuOpened((current) => !current)}
              >
                <img src={avatar} alt="" className={classes.AvatarIcon} />
                <span className={`material-symbols-outlined ${classes.UserArrow} ${
                  isUserMenuOpened ? classes.UserArrowOpened : ""
                }`}>
                  keyboard_arrow_down
                </span>
              </button>
              {isUserMenuOpened && (
                <div className={classes.LogoutMenu}>
                  <button
                    type="button"
                    className={classes.LogoutButton}
                    onClick={handleLogout}
                    disabled={isLoggingOut}
                  >
                    {isLoggingOut ? "Выход..." : "Выйти"}
                  </button>
                </div>
              )}
            </div>
          )}
        </Group>
      </AppShell.Header>
      <AppShell.Main className={classes.main}>
         <Outlet />
      </AppShell.Main>
    </AppShell>
  );
}
