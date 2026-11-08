import { SimpleGrid, Pagination, Group, Button, Modal, Text, TextInput, Popover, Radio, SegmentedControl, Select, Stack } from "@mantine/core";
import { CreateReportCard } from "../../components/CreateReportCard/CreateReportCard";
import { ReportCard } from "../../components/ReportCard/ReportCard";
import reportFormClasses from "../../components/CreateReportCard/CreateReportCard.module.css";

import { useDisclosure } from "@mantine/hooks";
import { useEffect, useMemo, useState } from "react";

import emptyReports from "../../assets/EmptyReports.svg";
import classes from "./ProjectPage.module.css";

const API_BASE_URL = import.meta.env.VITE_API_URL ?? "";

export function ProjectsPage(){

    const [opened, { open, close }] = useDisclosure(false);
    const [page, setPage] = useState(1);
    const [reports, setReports] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [apiError, setApiError] = useState(null);
    const [searchQuery, setSearchQuery] = useState("");
    const [salaryFilter, setSalaryFilter] = useState("all");
    const [quarterFilter, setQuarterFilter] = useState("");
    const [orgFilter, setOrgFilter] = useState("");
    const [appliedSalaryFilter, setAppliedSalaryFilter] = useState("all");
    const [appliedQuarterFilter, setAppliedQuarterFilter] = useState("");
    const [appliedOrgFilter, setAppliedOrgFilter] = useState("");
    const [deletingReportId, setDeletingReportId] = useState(null);
    const [deleteModalOpened, { open: openDeleteModal, close: closeDeleteModal }] = useDisclosure(false);
    const [reportToDeleteId, setReportToDeleteId] = useState(null);

    const reportsPerPage = 6;
    const isFilterActive = appliedSalaryFilter !== "all" || appliedQuarterFilter !== "" || appliedOrgFilter !== "";

    function applyFilters(){
        setAppliedSalaryFilter(salaryFilter);
        setAppliedQuarterFilter(quarterFilter);
        setAppliedOrgFilter(orgFilter);
        setPage(1);
    }

    function resetFilters(){
        setSalaryFilter("all");
        setQuarterFilter("");
        setOrgFilter("");
        setAppliedSalaryFilter("all");
        setAppliedQuarterFilter("");
        setAppliedOrgFilter("");
        setPage(1);
    }

    function handleDeleteRequest(reportId){
        if (!reportId){
            return false;
        }

        setReportToDeleteId(reportId);
        openDeleteModal();
    }

    function handleDeleteCancel(){
        if (deletingReportId){
            return;
        }

        setReportToDeleteId(null);
        closeDeleteModal();
    }

    async function handleDeleteConfirm(){
        if (!reportToDeleteId){
            return;
        }

        const isDeleted = await handleDeleteReport(reportToDeleteId);

        if (isDeleted){
            setReportToDeleteId(null);
            closeDeleteModal();
        }
    }

    async function handleDeleteReport(reportId){
        if (!reportId){
            return false;
        }

        try{
            setApiError(null);
            setDeletingReportId(reportId);

            const response = await fetch(
                `${API_BASE_URL}/api/v1/calculations/${reportId}`,
                {
                    method: "DELETE",
                    credentials: "include",
                }
            );

            let responseBody = {};
            try{
                responseBody = await response.json();
            }catch{
                responseBody = {};
            }

            if (!response.ok){
                throw new Error(responseBody?.error?.message ?? `Ошибка удаления: ${response.status}`);
            }

            setReports((currentReports) =>
                currentReports.filter((report) => report.id !== reportId)
            );
            return true;
        }catch(error){
            setApiError(error.message || "Не удалось удалить отчёт");
            return false;
        }finally{
            setDeletingReportId(null);
        }
    }

    const filteredReports = useMemo(() => {
        const query = searchQuery.trim().toLowerCase();

        return reports.filter((report) => {
        const text = `${report?.title ?? ""} ${report?.org_name ?? ""}`
        .toLowerCase();

        const matchesSearch = text.includes(query);

        const matchesSalary = appliedSalaryFilter === "all" || (appliedSalaryFilter === "pass" && report?.verdict === "PASS") ||
        (appliedSalaryFilter === "fail" && report?.verdict === "FAIL")

        const matchesQuarter = !appliedQuarterFilter || String(report?.period?.quarter) === appliedQuarterFilter;

        const matchesOrg =
        !appliedOrgFilter ||
        report?.org_name === appliedOrgFilter;

        return matchesSearch && matchesSalary && matchesQuarter && matchesOrg;
        });
    }, [reports, searchQuery, appliedSalaryFilter, appliedQuarterFilter, appliedOrgFilter]);
    const totalPages = Math.max(1, Math.ceil(filteredReports.length / reportsPerPage));

    useEffect(() => {
        if (page > totalPages){
            setPage(totalPages);
        }
    }, [page, totalPages]);

    useEffect(() => {
        let isMounted = true;

        async function loadReports(){
            try{
                setIsLoading(true);
                setApiError(null);
                const response = await fetch(`${API_BASE_URL}/api/v1/calculations`, {
                    credentials: "include",
                });
                let responseBody = {};
                try{
                    responseBody = await response.json();
                }catch{
                    responseBody = {};
                }
                if (!response.ok){
                    throw new Error(responseBody?.error?.message ?? `Ошибка запроса: ${response.status}`);
                }
                if (isMounted){
                    setReports(Array.isArray(responseBody.items) ? responseBody.items : []);
                    setPage(1);
                }
            }catch(error){
                if (isMounted){
                    setApiError(error.message || "Не удалось загрузить отчёты");
                }
            }finally{
                if (isMounted){
                    setIsLoading(false);
                }
            }
        }
        loadReports();
        return () => {
            isMounted = false;
        };
    }, []);

    const visibleReports = useMemo(() => filteredReports.slice(
        (page - 1) * reportsPerPage,
        page * reportsPerPage
    ), [page, filteredReports]);

    const organizationOptions = useMemo(() =>{
        return [...new Set(
            reports.map((report) => report?.org_name).filter(Boolean)
        )];
    }, [reports]);

    return(
        <section className={classes.page}>
            <Group justify="space-between">
                <Group gap={16}>
                <TextInput 
                placeholder="Поиск"
                className={classes.Search}
                value={searchQuery}
                onChange={
                    (event) => {
                        setSearchQuery(event.currentTarget.value);
                        setPage(1);
                    }
                }
                rightSection={
                   <span className="material-symbols-outlined">
                        search
                   </span> 
                }
                />
                <Popover width={470} position="bottom-start" shadow="md" radius={16}>
                <Popover.Target>
                    <Button
                        className={`${classes.ButtonFilter} ${isFilterActive ? classes.ButtonFilterActive : ""}`}
                        classNames={{
                            inner: classes.ButtonFilterInner,
                            section: classes.ButtonFilterSection,
                        }}
                        rightSection={
                            <span className="material-symbols-outlined">
                                filter_list
                            </span>
                        }
                    >
                        Фильтры
                    </Button>
                </Popover.Target>
                <Popover.Dropdown className={classes.Card}>
                    <Stack gap={24}>
                        <Stack gap={4}>
                        <Text className={classes.TextButtonFilter}>
                            Критерии по заработной плате
                        </Text>
                        <Radio.Group value={salaryFilter} onChange={setSalaryFilter}>
                            <Group gap={12}>
                            <Radio 
                            value="all" 
                            label="Все" 
                            classNames={{
                                radio: classes.RadioInput,
                                label: classes.RadioLabel,
                                }}
                            />
                            <Radio 
                            value="pass" 
                            label="Выполняется" 
                            classNames={{
                                radio: classes.RadioInput,
                                label: classes.RadioLabel,
                                }}
                            />
                            <Radio 
                            value="fail" 
                            label="Не выполняется" 
                            classNames={{
                                radio: classes.RadioInput,
                                label: classes.RadioLabel,
                                }}
                            />
                            </Group>
                        </Radio.Group>
                        </Stack>
                        <Stack gap={8}>
                        <Text className={classes.TextButtonFilter}>
                            Квартал
                        </Text>
                        <SegmentedControl
                        value={quarterFilter} 
                        onChange={setQuarterFilter}
                        withItemsBorders={false}
                        classNames={{
                            root: classes.QuarterRoot,
                            label: classes.QuarterLabel,
                            indicator: classes.QuarterIndicator,
                            control: classes.QuarterControl,
                        }}
                        data={[
                            { label: "I", value: "1" },
                            { label: "II", value: "2" },
                            { label: "III", value: "3" },
                            { label: "IV", value: "4" },
                            ]}
                        />
                        </Stack>
                            <Stack gap={8}>
                                <Text className={classes.TextButtonFilter}>
                                    Организация
                                </Text>
                                <Select
                                    value={orgFilter} 
                                    onChange={(value) => setOrgFilter(value ?? "")}
                                    className={classes.Select}
                                    placeholder="Выберите организацию"
                                    data={organizationOptions}
                                    comboboxProps={{ withinPortal: false }}
                                />
                                <Group justify="flex-end">
                                <Button className={classes.ButtonAgreeFilter} onClick={applyFilters}>
                                    Применить
                                </Button>
                                </Group>
                            </Stack>
                    </Stack>
                </Popover.Dropdown>
                </Popover>
                {isFilterActive && (
                <Button
                variant="subtle"
                className={classes.ResetFilters}
                onClick={resetFilters}
                rightSection={
                    <span className="material-symbols-outlined">
                        close
                    </span>
                }
                >
                    Сбросить все
                </Button>
                )}
                </Group>
                <Button className={classes.Button} onClick={open}>
                    Создать новый отчёт
                </Button>
            </Group>
            <Modal
                opened={opened}
                onClose={close}
                centered
                size="calc((100vw - 152px) / 3)"
                radius={16}
                overlayProps={{
                    backgroundOpacity: 0.15,
                    blur: 0,
                }}
                classNames={{
                    content: reportFormClasses.ModalContent,
                    body: reportFormClasses.ModalBody,
                    header: reportFormClasses.ModalHeader
                }}
            >
                <CreateReportCard/>
            </Modal>
            <Modal
                opened={deleteModalOpened}
                onClose={handleDeleteCancel}
                centered
                withCloseButton={false}
                closeOnClickOutside={!deletingReportId}
                closeOnEscape={!deletingReportId}
                radius={36}
                size={429}
                overlayProps={{
                    backgroundOpacity: 0.15,
                    blur: 0,
                }}
                classNames={{
                    content: classes.DeleteModalContent,
                    body: classes.DeleteModalBody,
                }}
            >
                <button
                    type="button"
                    className={classes.DeleteModalClose}
                    onClick={handleDeleteCancel}
                    disabled={Boolean(deletingReportId)}
                    aria-label="Закрыть"
                >
                    <span className="material-symbols-outlined">
                        close
                    </span>
                </button>
                <Stack gap={0}>
                    <Text className={classes.DeleteModalTitle}>
                        Вы уверены, что хотите удалить отчет?
                    </Text>
                    <Text className={classes.DeleteModalText}>
                        Вы не сможете восстановить данные
                    </Text>
                    <Group grow className={classes.DeleteModalActions}>
                        <Button
                            className={classes.DeleteCancelButton}
                            onClick={handleDeleteCancel}
                            disabled={Boolean(deletingReportId)}
                        >
                            Отменить
                        </Button>
                        <Button
                            className={classes.DeleteConfirmButton}
                            onClick={handleDeleteConfirm}
                            loading={Boolean(deletingReportId)}
                        >
                            Удалить
                        </Button>
                    </Group>
                </Stack>
            </Modal>
            {apiError && (
                <Text c="red" mt={16}>
                    {apiError}
                </Text>
            )}
            {isLoading && (
                <Text c="dimmed" mt={16}>
                    Загрузка отчётов...
                </Text>
            )}
            {!isLoading && !apiError && filteredReports.length === 0 && (
                <div className={classes.EmptyState}>
                    <img src={emptyReports} alt="" className={classes.EmptyImage} />
                    <Text className={classes.EmptyText}>
                        {reports.length === 0
                            ? "У вас пока нет отчетов, создайте новый отчет"
                            : "По выбранным параметрам ничего не найдено"}
                    </Text>
                </div>
            )}
            {!isLoading && !apiError && filteredReports.length > 0 && (
            <SimpleGrid cols={3} spacing={16} className={classes.simpleGrid}>
                    {visibleReports.map((report) => (
                        <ReportCard
                        key={report.id}
                        report={report}
                        onDelete={handleDeleteRequest}
                        isDeleting={deletingReportId === report.id}
            />
            ))}
            </SimpleGrid>
            )}
            <div>
                {filteredReports.length > reportsPerPage && (
                <Pagination 
                value={page}
                onChange={setPage}
                size="lg" 
                className={classes.pagination}

                classNames={{
                    control: classes.paginationControl
                }}
                color="dark" 
                withControls={false}
                total={totalPages}
                 />    
                )}
            </div>
        </section>
    );
}
